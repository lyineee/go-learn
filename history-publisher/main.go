package main

import (
	"context"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/lyineee/go-learn/utils"
	"github.com/lyineee/go-learn/utils/log"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/zap"
)

type RedisQueueOptions struct {
	Group  string
	Stream string
}
type RedisStreamMessage struct {
	ID        string
	MongoDBId string
}
type History struct {
	Id        primitive.ObjectID `bson:"_id,omitempty"`
	Url       string             `bson:"url,omitempty"`
	Type      string             `bson:"type,omitempty"`
	TotalPage int                `bson:"total_page,omitempty"`
	Title     string             `bson:"title,omitempty"`
}

var logger *log.SugarLogger

var historyDatabase string = "site"
var historyCol string = "history"

func main() {
	// get envirment
	envMap := utils.GetEnv()
	redisAddress, ok := envMap["REDIS"]
	if !ok {
		redisAddress = "redis:6379"
		log.Default().Sugar().Warn("Use default redis server address", "redis", redisAddress)
	}

	mongoAddress, ok := envMap["MONGODB"]
	if !ok {
		mongoAddress = "mongodb://mongodb:27017/"
		log.Default().Sugar().Warn("Use default mongo server address", "mongo_address", mongoAddress)

	}
	logStream, ok := envMap["LOG_STREAM"]
	if !ok {
		log.Fatal("no env LOG_STREAM")
	}
	w := log.NewRedisWriterWithAddress(redisAddress, "", logStream, "go-learn.history-publisher")
	logger = log.NewLogger(log.NewJsonCore(w), log.InfoLevel).Sugar()
	defer logger.Sync()

	// init redis db
	var ctx = context.Background()
	rdb := redis.NewClient(&redis.Options{
		Addr:     redisAddress,
		Password: "", // no password set
		DB:       0,  // use default DB
	})

	//init mongodb
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	mongoClient, err := mongo.Connect(ctx, options.Client().ApplyURI(mongoAddress))
	if err != nil {
		log.Panic("Fail connect to MongoDB", zap.Any("err", err))
	}

	redisQueueOptions := RedisQueueOptions{
		Group:  "backend.history.refresh.workers",
		Stream: "backend.history.refresh",
	}
	//TODO create group worker
	err = createGroup(ctx, rdb, redisQueueOptions)
	if err != nil {
		logger.Errorw("create group fail", "queue_options", redisQueueOptions, "error", err)
	}

	//TODO clean up stream pending list

	historyCol := mongoClient.Database(historyDatabase).Collection(historyCol)
	historys, err := getAllHistory(ctx, historyCol)
	if err != nil {
		logger.Errorw("get history from mongodb error", "error", err) //TODO add handler
	}
	for _, history := range historys {
		//TODO check if stream exist
		err = addIdToStream(ctx, rdb, redisQueueOptions, history)
		logger.Infow(fmt.Sprintf("add history %s to stream", history.Id.Hex()), "history", history)
		if err != nil {
			logger.Errorw("add history to stream error", "redis_options", redisQueueOptions, "history", history, "error", err)
		}
	}
}

func getAllHistory(ctx context.Context, col *mongo.Collection) (historys []History, err error) {
	//TODO validate owner account status
	filter := bson.M{"$or": bson.A{bson.M{"delete": false}, bson.M{"delete": bson.M{"$exists": false}}}}
	cur, err := col.Find(ctx, filter)
	if err != nil {
		return
	}
	err = cur.All(ctx, &historys)
	if err != nil {
		logger.Errorw("get history error", "error", err) //TODO add error handler
	}
	return
}

func addIdToStream(ctx context.Context, rdb *redis.Client, options RedisQueueOptions, history History) error {
	// for _, history := range historys {
	result, err := rdb.XAdd(ctx, &redis.XAddArgs{
		Stream:     options.Stream,
		NoMkStream: false,
		ID:         "*",
		Values: map[string]interface{}{
			"id": history.Id.Hex(),
		},
	}).Result()
	if err != nil {
		logger.Error("add to stream error", "error", err) //TODO add error handler
	}
	if result == "0" { //TODO validate result value
		logger.Warnw("result is 0")
	}
	// }
	return nil

}

func createGroup(ctx context.Context, rdb *redis.Client, options RedisQueueOptions) error {
	groupInfos, err := rdb.XInfoGroups(ctx, options.Stream).Result()
	if err != nil {
		return err
	}
	for _, info := range groupInfos {
		if info.Name == options.Group {
			return nil
		}
	}
	logger.Infow(fmt.Sprintf("group %s not found in stream %s", options.Group, options.Stream))
	result, err := rdb.XGroupCreate(ctx, options.Stream, options.Group, "$").Result()
	if err != nil {
		logger.Errorw("error when create group", "group_options", options, "error", err)
		return err
	}
	logger.Infow("create group", "group_option", options, "return_code", result)
	return nil
}
