package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/lyineee/go-learn/utils/log"
	_ "github.com/lyineee/go-learn/utils/remote"
	"github.com/spf13/viper"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
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
var logSubject string = "go-learn.history-publisher"

func main() {
	// get envirment
	viper.AutomaticEnv()

	// init etcd config
	//viper.SetDefault("etcd", "etcd:2379")
	viper.SetDefault("etcd_config_path", "/config/history-publisher.toml")

	//database
	viper.SetDefault("database.redis", "redis:6379")
	viper.SetDefault("database.mongo", "mongodb://mongodb:27017")

	//redis stream
	viper.SetDefault("log.subject", logSubject)

	log.Info("default config", log.Any("config", viper.AllSettings()))

	if viper.IsSet("etcd") {
		viper.AddRemoteProvider("etcd", viper.GetString("etcd"), viper.GetString("etcd_config_path"))
		viper.SetConfigType("toml")
		err := viper.ReadRemoteConfig()
		if err != nil {
			log.Panic("cannot connect to etcd config center", log.String("etcd", viper.GetString("etcd")), log.String("etcd_path", viper.GetString("etcd_config_path")), log.Error(err))
		}
	} else {
		viper.SetConfigName("history-publisher.toml")
		viper.SetConfigType("toml")
		viper.AddConfigPath("/etc/")
		viper.AddConfigPath("./app")
		viper.AddConfigPath("./")
		if err := viper.ReadInConfig(); err != nil {
			if _, ok := err.(viper.ConfigFileNotFoundError); ok {
				log.Fatal("Config file not found")
			} else {
				log.Fatal("Error when read config", log.Error(err))
			}
		}
	}
	log.Info("all config", log.Any("config", viper.AllSettings()))

	if !viper.IsSet("log.stream") {
		log.Info("no stream name, using stdout")
		w := os.Stdout
		logger = log.NewLogger(log.NewJsonCore(w), log.InfoLevel).Sugar()
	} else {
		subject := viper.GetString("log.subject")
		logStream := viper.GetString("log.stream")

		log.Info("using redis log stream", log.String("log_stream", logStream), log.String("log_subject", subject))
		w := log.NewRedisWriterWithAddress(viper.GetString("database.redis"), "", logStream, subject)
		logger = log.NewLogger(log.NewJsonCore(w), log.InfoLevel).Sugar()
	}
	defer logger.Sync()

	// init redis db
	var ctx = context.Background()
	rdb := redis.NewClient(&redis.Options{
		Addr:     viper.GetString("database.redis"),
		Password: viper.GetString("database.redis_password"), // no password set
		DB:       viper.GetInt("database.redis_db"),          // use default DB
	})

	//init mongodb
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	mongoClient, err := mongo.Connect(ctx, options.Client().ApplyURI(viper.GetString("database.mongo")))
	if err != nil {
		log.Panic("Fail connect to MongoDB", log.Any("err", err))
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
