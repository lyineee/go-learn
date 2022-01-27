package main

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"
	"strconv"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
	"github.com/lyineee/go-learn/utils"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/zap"
	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/transform"
)

type postRequestFunc func(client *http.Client, crawlUrl string) (*http.Request, error)

type NgaPostInformation struct {
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
		logger.Warn("Use default mongo server address", "mongodb", mongoAddress)

	}

	var ctxBackground = context.Background()
	// init redis db
	rdb := redis.NewClient(&redis.Options{
		Addr:     redisAddress,
		Password: "", // no password set
		DB:       0,  // use default DB
	})

	//init mongodb
	ctxMongoConnect, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	mongoClient, err := mongo.Connect(ctxMongoConnect, options.Client().ApplyURI(mongoAddress))
	if err != nil {
		log.Panic("Fail connect to MongoDB", zap.Any("err", err))
	}
	ctxPing, cancelPing := context.WithTimeout(ctxBackground, 10*time.Second)
	defer cancelPing()
	err = mongoClient.Ping(ctxPing, nil)
	if err != nil {
		logger.Fatal(fmt.Sprintf("connect to mongodb %s fail, exit", mongoAddress))
	}

	uuid := uuid.NewString()
	redisQueueOptions := RedisQueueOptions{
		Group:      "backend.history.refresh.workers",
		Stream:     "backend.history.refresh",
		ComsumerID: uuid,
	}

	logger.Info("new consumer", "group", redisQueueOptions.Group, "comsumer_id", redisQueueOptions.ComsumerID)

	for {
		msg, err := claimMessage(ctxBackground, rdb, redisQueueOptions)
		ctxTimeout, cancelContext := context.WithTimeout(ctxBackground, 20*time.Second)
		if err != nil {
			continue
		}
		historyCol := mongoClient.Database(historyDatabase).Collection(historyCol)
		history, err := getHistory(ctxTimeout, historyCol, msg.MongoDBId)
		if err != nil {
			logger.Error("error", err) //TODO error handler
			continue
		}
		logger.Info("get history", "history", history, "consumer_id", redisQueueOptions.ComsumerID)
		switch history.Type {
		case "nga":
		default: //TODO only nga, for test
			err := ngaProc(ctxTimeout, &history)
			if err != nil {
				logger.Error("error", err)
				continue
			}
		}
		logger.Info("complete process", "history", history, "consumer_id", redisQueueOptions.ComsumerID)
		err = updateHistory(ctxTimeout, historyCol, history)
		if err != nil {
			logger.Error("error", err)
			continue
		}
		logger.Info("crawl success, group ack", "queue_id", msg.ID, "historyId", history.Id.Hex())
		ACKMessage(ctxTimeout, rdb, redisQueueOptions, msg)
		cancelContext()
	}
}

func claimMessage(ctx context.Context, rdb *redis.Client, options RedisQueueOptions) (msg RedisStreamMessage, err error) {
	logger.Info("waiting for group message", "stream", options.Stream, "group", options.Group)
	stream, err := rdb.XReadGroup(ctx, &redis.XReadGroupArgs{
		Group:    options.Group,
		Streams:  []string{options.Stream, ">"},
		Consumer: options.ComsumerID,
		Count:    1,
		Block:    0,
		NoAck:    false,
	}).Result()
	if err != nil {
		logger.Error("read redis group fail", "error", err)
		return
	}
	msg.ID = stream[0].Messages[0].ID
	for key, value := range stream[0].Messages[0].Values {
		switch key {
		case "id":
			msg.MongoDBId = value.(string)
		}
	}
	logger.Info("get message", "message", msg)
	return msg, nil
}

func ACKMessage(ctx context.Context, rdb *redis.Client, options RedisQueueOptions, msg RedisStreamMessage) error {

	result, err := rdb.XAck(ctx, options.Stream, options.Group, msg.ID).Result()
	if err != nil {
		logger.Error("fail to ack redis queue", "error", err)
		return err
	}
	if result == 0 {
		logger.Warn("ack already done", "message", msg)
	}
	return nil
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
	if err != nil {
		return err
	}
	history.Title = info.Title
	history.TotalPage = info.TotalPage
	return nil
}

func crawlPage(crawlUrl string, postRequest postRequestFunc) (string, error) {
	logger.Debug("start crwaling page", "url", crawlUrl)
	client := http.Client{}
	req, err := postRequest(&client, crawlUrl)
	if err != nil {
		logger.Error("Error when process postReqeust function", "crawl_url", crawlUrl, "error", err)
		return "", err
	}
	logger.Debug("get request", "cookies", req.Cookies())
	resp, err := client.Do(req)
	if err != nil {
		logger.Error("Error when request crawl url", "crawl_url", crawlUrl, "error", err)
		return "", err
	}
	defer resp.Body.Close()
	bytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		logger.Error("Error read response body", "crawl_url", crawlUrl, "error", err)
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

func ngaExtractor(text string) (information NgaPostInformation, err error) {
	//TODO more accurate way to detect charset
	text, err = GBKToUTF8(text)
	if err != nil {
		return information, errors.New("string encoding fail")
	}
	regTitle := regexp.MustCompile(`<title>(.+?)</title>`)
	titleRow := regTitle.FindStringSubmatch(text)
	if len(titleRow) < 2 {
		return information, errors.New("can not find title")
	}
	information.Title = titleRow[1]

	regTotlePage := regexp.MustCompile(`__PAGE.+?,\d+:(\d*)`)
	totlePageRow := regTotlePage.FindStringSubmatch(text)
	if len(totlePageRow) < 2 {
		return information, errors.New("can not find totle page")
	}
	totlePage, err := strconv.Atoi(totlePageRow[1])
	if err != nil {
		return information, errors.New("can not find totle page")
	}
	information.TotalPage = totlePage
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
