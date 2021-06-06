package main

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/lyineee/go-learn/utils"
	"go.uber.org/zap"
)

var logger *zap.SugaredLogger

func main() {
	// logger init
	log := utils.GetLogger()
	logger = log.Sugar()
	// logger = zap.NewExample().Sugar()
	logger.Info("Logger init finished")

	// get envirment
	envMap := utils.GetEnv()
	redisHost, ok1 := envMap["REDIS_HOST"]
	redisPort, ok2 := envMap["REDIS_PORT"]
	if !(ok1 || ok2) {
		logger.Warn("Use default redis server address")
		redisHost = "localhost"
		redisPort = "6379"
	}

	studentId, ok := envMap["STUDENT_ID"]
	if !ok {
		logger.Panic("Dont find STUDENT_ID")
	}
	logger.Infof("Get environment STUDENT_ID: %v", studentId)

	cardNo, ok := envMap["CARD_NO"]
	if !ok {
		logger.Panic("Dont find CARD_NO")
	}
	logger.Infof("Get environment CARD_NO: %v", cardNo)

	apiKey, ok := envMap["XXBH"]
	if !ok {
		logger.Panic("Dont find apikey \"xxbk\"")
	}
	logger.Infof("Get apikey \"xxbh\": %v", apiKey)

	// redis init
	var ctx = context.Background()
	rdb := redis.NewClient(&redis.Options{
		Addr:     redisHost + ":" + redisPort,
		Password: "", // no password set
		DB:       0,  // use default DB
	})
	today := quiryToday(studentId, cardNo, apiKey)
	logger.Infof("today: %.2f", today)
	balance := quiryBanlance(studentId, cardNo, apiKey)
	logger.Infof("balance: %.2f", balance)
	rdb.Set(ctx, "test", balance, 0)
	cmd := rdb.HSet(ctx, studentId, map[string]interface{}{
		"balance":   balance,
		"today":     today,
		"timestamp": float32(time.Now().Unix()),
	})
	logger.Infof("Redis result: %v", cmd)
}

func quiryBanlance(studentId string, cardNo string, apiKey string) float32 {
	payload := url.Values{"sno": {studentId}, "xxbh": {apiKey}, "idtype": {"acc"}, "id": {cardNo}}
	resp, err := http.PostForm(balance_url, payload)
	if err != nil {
		logger.Panicf("Fail getting balance for studentId:%v cardNo:%v", studentId, cardNo, "err", err)
	}
	defer resp.Body.Close()
	var respJson HisResp
	if err := json.NewDecoder(resp.Body).Decode(&respJson); err != nil {
		logger.Error("Fail decoding json", "err", err)
		bodyByte, _ := ioutil.ReadAll(resp.Body)
		logger.Debugf("Response content is:%v", string(bodyByte))
	}
	if respJson.Retcode != "0" {
		logger.Panicf("Error getting balance with cardNo:%v", cardNo, "msg", respJson.Errmsg)
	}
	balance, err := strconv.ParseInt(respJson.Card[0].DbBalance, 10, 64)
	if err != nil {
		logger.Panic("Error converting balance from string to int", "respJson", respJson, "err", err)
	}
	return float32(balance) / 100
}

func quiryToday(studentId string, cardNo string, apiKey string) float32 {
	dateStr := time.Now().Format("20060102")
	payload := url.Values{
		"sno":         {studentId},
		"xxbh":        {apiKey},
		"idtype":      {"acc"},
		"id":          {cardNo},
		"curpage":     {"1"},
		"pagesize":    {"10"},
		"account":     {cardNo},
		"acctype":     {""},
		"query_start": {dateStr},
		"query_end":   {dateStr},
	}
	var total int = 0
	for index := 1; true; index++ {
		logger.Debugf("Current page: %v", index)
		payload.Set("curpage", strconv.Itoa(index))
		resp, err := http.PostForm(history_url, payload)
		if err != nil {
			logger.Panicf("Fail getting balance for studentId:%v cardNo:%v", studentId, cardNo, "err", err)
		}
		defer resp.Body.Close()
		var respJson TodayResp
		if err := json.NewDecoder(resp.Body).Decode(&respJson); err != nil {
			logger.Error("Fail decoding json", "err", err)
			bodyByte, _ := ioutil.ReadAll(resp.Body)
			logger.Debugf("Response content is:%v", string(bodyByte))
		}
		if respJson.Retcode != "0" {
			logger.Panicf("Error getting balance with cardNo:%v", cardNo, "msg", respJson.Errmsg)
		}
		for _, item := range respJson.Total {
			if strings.Contains(item.SignTranamt, "-") {
				result, err := strconv.ParseInt(item.Tranamt, 10, 64)
				if err != nil {
					logger.Error("Fail in parse today history", "Tranamt", item.Tranamt, "err", err)
				} else {
					total += int(result)
				}
			}
		}
		if respJson.NextPage == "0" {
			break
		}
	}
	return float32(total) / 100
}
