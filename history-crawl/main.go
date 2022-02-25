package main

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"time"

	"github.com/lyineee/go-learn/utils/log"
	_ "github.com/lyineee/go-learn/utils/remote"
	"github.com/spf13/viper"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/transform"

	rstream "github.com/lyineee/go-learn/redis-stream"
)

type postRequestFunc func(client *http.Client, crawlUrl string) (*http.Request, error)

type postInformation struct {
	Title     string
	TotalPage int
}

type RedisQueueOptions struct {
	Group      string
	Stream     string
	ComsumerID string
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
var logSubject string = "go-learn.history-crawl"

func main() {
	// get envirment
	viper.AutomaticEnv()

	// init etcd config
	viper.SetDefault("etcd", "etcd:2379")
	viper.SetDefault("etcd_config_path", "/config/history-crawl.toml")

	//database
	viper.SetDefault("database.redis", "redis:6379")
	viper.SetDefault("database.mongo", "mongodb://mongodb:27017")

	//redis stream
	viper.SetDefault("log.subject", logSubject) //log
	viper.SetDefault("stream.group", "backend.history.refresh.workers")
	viper.SetDefault("stream.stream", "backend.history.refresh")

	viper.AddRemoteProvider("etcd", viper.GetString("etcd"), viper.GetString("etcd_config_path"))
	viper.SetConfigType("toml")
	err := viper.ReadRemoteConfig() //get remote config
	if err != nil {
		log.Panic("cannot connect to etcd config center", log.String("etcd", viper.GetString("etcd")), log.String("etcd_path", viper.GetString("etcd_config_path")), log.Error(err))
	}
	log.Info("all config", log.Any("config", viper.AllSettings())) //print all config

	//init logger
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

	var ctxBackground = context.Background()
	redisQueueOptions := RedisQueueOptions{
		Group:  viper.GetString("stream.group"),
		Stream: viper.GetString("stream.stream"),
	}
	group, err := rstream.NewGroup(ctxBackground, viper.GetString("database.redis"), "", redisQueueOptions.Group, redisQueueOptions.Stream)
	if err != nil {
		logger.Fatalw("fail connect to redis", "redis_config", redisQueueOptions, "error", err)
	}
	redisQueueOptions.ComsumerID = group.ConsumerID

	//init mongodb
	ctxMongoConnect, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	mongoClient, err := mongo.Connect(ctxMongoConnect, options.Client().ApplyURI(viper.GetString("database.mongo")))
	if err != nil {
		logger.Panicw("Fail connect to MongoDB", "err", err)
	}
	ctxPing, cancelPing := context.WithTimeout(ctxBackground, 10*time.Second)
	defer cancelPing()
	err = mongoClient.Ping(ctxPing, nil)
	if err != nil {
		logger.Fatalw(fmt.Sprintf("connect to mongodb %s fail, exit", viper.GetString("database.mongo")))
	}

	if err != nil {
		logger.Fatalw("fail init redis stream", "error", err)
	}

	logger.Infow("new consumer", "group", redisQueueOptions.Group, "comsumer_id", redisQueueOptions.ComsumerID)

	for {
		msg, err := claimMessage(ctxBackground, group)
		ctxTimeout, cancelContext := context.WithTimeout(ctxBackground, 20*time.Second)
		if err != nil {
			continue
		}
		historyCol := mongoClient.Database(historyDatabase).Collection(historyCol)
		history, err := getHistory(ctxTimeout, historyCol, msg.MongoDBId)
		if err != nil {
			logger.Errorw("get history error", "error", err) //TODO error handler
			continue
		}
		logger.Infow("get history", "history", history, "consumer_id", redisQueueOptions.ComsumerID)
		switch history.Type {
		case "nga":
			err := ngaProc(ctxTimeout, &history)
			if err != nil {
				logger.Errorw("process nga error", "error", err)
				continue
			}
		case "tieba":
			err := tiebaProc(ctxTimeout, &history)
			if err != nil {
				logger.Errorw("process tieba error", "error", err)
				continue
			}
		default:
			logger.Errorw("ack with no extractor", "history", history, "type", history.Type)
			if err := group.Ack(ctxTimeout, msg.ID); err != nil {
				logger.Errorw("ack error", "history", history, "queue_msg", msg)
			}
		}
		logger.Infow("complete process", "history", history, "consumer_id", redisQueueOptions.ComsumerID)
		err = updateHistory(ctxTimeout, historyCol, history)
		if err != nil {
			logger.Error("mongodb update history error", "error", err)
			continue
		}
		logger.Infow("crawl success, group ack", "queue_id", msg.ID, "historyId", history.Id.Hex())
		if err := group.Ack(ctxTimeout, msg.ID); err != nil {
			logger.Errorw("ack error", "history", history, "queue_msg", msg)
		}
		cancelContext()
	}
}

func claimMessage(ctx context.Context, group rstream.ConsumerGroup) (msg RedisStreamMessage, err error) {
	message, err := group.Pop()
	if err != nil {
		return
	}
	msg.ID = message.ID
	for key, value := range message.Values {
		switch key {
		case "id":
			msg.MongoDBId = value.(string)
		}
	}
	logger.Infow("get message", "message", msg)
	return msg, nil
}

func getHistory(ctx context.Context, col *mongo.Collection, historyId string) (history History, err error) {
	id, err := primitive.ObjectIDFromHex(historyId)
	if err != nil {
		return history, err
	}
	filter := bson.M{"_id": id}
	result := col.FindOne(ctx, filter)
	err = result.Decode(&history)
	if err != nil {
		return history, err
	}
	return history, nil
}

func updateHistory(ctx context.Context, col *mongo.Collection, history History) error {
	filter := bson.M{"_id": history.Id}
	update := bson.M{"$set": bson.M{"title": history.Title, "total_page": history.TotalPage}}
	_, err := col.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}
	return nil
}

func ngaProc(ctx context.Context, history *History) error {
	page, err := crawlPage(history.Url, ngaPostRequest)
	if err != nil {
		return err
	}
	info, err := ngaExtractor(page)
	if info.Title == "" && info.TotalPage == 0 {
		logger.Errorw("get nga info fail", "crawl page", page, "history", history)
		return errors.New("get info fail")
	} else if info.Title == "" || info.TotalPage == 0 {
		logger.Warnw("fail get all nga data", "crawl page", page, "history", history)
	}
	if err != nil {
		return err
	}
	history.Title = info.Title
	history.TotalPage = info.TotalPage
	return nil
}

func crawlPage(crawlUrl string, postRequest postRequestFunc) (string, error) {
	logger.Debugw("start crwaling page", "url", crawlUrl)
	client := http.Client{}
	req, err := postRequest(&client, crawlUrl)
	if err != nil {
		logger.Errorw("Error when process postReqeust function", "crawl_url", crawlUrl, "error", err)
		return "", err
	}
	logger.Debugw("get request", "cookies", req.Cookies())
	resp, err := client.Do(req)
	if err != nil {
		logger.Errorw("Error when request crawl url", "crawl_url", crawlUrl, "error", err)
		return "", err
	}
	defer resp.Body.Close()
	bytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		logger.Errorw("Error read response body", "crawl_url", crawlUrl, "error", err)
	}
	return string(bytes), nil
}

func ngaPostRequest(client *http.Client, crawlUrl string) (*http.Request, error) {
	resp, err := client.Get(crawlUrl)
	if err != nil {
		return nil, err
	}
	bytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	respText := string(bytes)
	regGuestJs := regexp.MustCompile(`guestJs=(\d+)`)
	guestJsRow := regGuestJs.FindStringSubmatch(respText)
	if len(guestJsRow) < 2 {
		return nil, errors.New("can not find guestJs")
	}
	guestJs := guestJsRow[1]
	req, err := http.NewRequest("GET", crawlUrl, nil)
	if err != nil {
		return nil, err
	}

	cookie := http.Cookie{
		Name:    "guestJs",
		Value:   guestJs,
		Expires: time.Now().Local().Add(time.Second * 12000),
		Domain:  "bbs.nga.cn",
	}
	req.AddCookie(&cookie)
	time.Sleep(time.Second * 1)
	return req, nil
}

func ngaExtractor(text string) (information postInformation, err error) {
	//TODO more accurate way to detect charset
	text, err = GBKToUTF8(text)
	if err != nil {
		return information, errors.New("string encoding fail")
	}
	regTitle := regexp.MustCompile(`<title>(.+?)</title>`)
	titleRaw := regTitle.FindStringSubmatch(text)
	if len(titleRaw) < 2 {
		return information, errors.New("can not find title")
	}
	information.Title = titleRaw[1]

	regTotalPage := regexp.MustCompile(`__PAGE.+?,\d+:(\d*)`)
	totalPageRaw := regTotalPage.FindStringSubmatch(text)
	if len(totalPageRaw) < 2 {
		return information, errors.New("can not find total page")
	}
	totalPage, err := strconv.Atoi(totalPageRaw[1])
	if err != nil {
		return information, errors.New("can not find total page")
	}
	information.TotalPage = totalPage
	return information, nil
}

func GBKToUTF8(s string) (string, error) {
	reader := transform.NewReader(bytes.NewReader([]byte(s)), simplifiedchinese.GBK.NewDecoder())
	d, err := ioutil.ReadAll(reader)
	if err != nil {
		return "", err
	}
	return string(d), nil
}
