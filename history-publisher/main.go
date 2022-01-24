package main

import (
	"context"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/lyineee/go-learn/utils"
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

var logger *zap.SugaredLogger

var historyDatabase string = "site"
var historyCol string = "history"

func main() {
	log := utils.GetLogger()
	logger = log.Sugar()
	defer logger.Sync()

	// get envirment
	envMap := utils.GetEnv()
	redisAddress, ok := envMap["REDIS"]
	if !ok {
		redisAddress = "redis:6379"
		logger.Warn("Use default redis server address", "redis", redisAddress)
	}

	mongoAddress, ok := envMap["MONGODB"]
	if !ok {
		mongoAddress = "mongodb://mongodb:27017/"
		logger.Warn("Use default mongo server address", "mongo_address", mongoAddress)

	}

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
		logger.Error("create group fail", "queue_options", redisQueueOptions, "error", err)
	}

	//TODO clean up stream pending list

	historyCol := mongoClient.Database(historyDatabase).Collection(historyCol)
	historys, err := getAllHistory(ctx, historyCol)
	if err != nil {
		logger.Error("error", err) //TODO add handler
	}
	for _, history := range historys {
		//TODO check if stream exist
		err = addIdToStream(ctx, rdb, redisQueueOptions, history)
		logger.Info(fmt.Sprintf("add history %s to stream", history.Id.Hex()), "history", history)
		if err != nil {
			logger.Error("error", err)
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
		logger.Error("error", err) //TODO add error handler
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
		logger.Error("error", err) //TODO add error handler
	}
	if result == "0" { //TODO validate result value
		logger.Warn("result is 0")
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
	logger.Info(fmt.Sprintf("group %s not found in stream %s", options.Group, options.Stream))
	result, err := rdb.XGroupCreate(ctx, options.Stream, options.Group, "$").Result()
	if err != nil {
		logger.Error("error when create group", "group_options", options, "error", err)
		return err
	}
	logger.Info("create group", "group_option", options, "return_code", result)
	return nil
}
