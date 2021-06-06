package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/lyineee/go-learn/utils"
	"go.uber.org/zap"
)

type ClassCode struct {
	Code string
	Name string
}

type ClassAllResp struct {
	E int
	M string
	D struct{ All map[string][]ClassCode }
}

type ClassJsResp struct {
	E int
	M string
	D struct{ Js map[string][]ClassCode }
}

var logger *zap.SugaredLogger

func main() {
	log := utils.GetLogger()
	logger = log.Sugar()
	defer logger.Sync()

	// get envirment
	envMap := utils.GetEnv()
	redisHost, ok1 := envMap["REDIS_HOST"]
	redisPort, ok2 := envMap["REDIS_PORT"]
	if !(ok1 || ok2) {
		logger.Warn("Use default redis server address")
		redisHost = "localhost"
		redisPort = "6379"
	}

	// init redis db
	var ctx = context.Background()
	rdb := redis.NewClient(&redis.Options{
		Addr:     redisHost + ":" + redisPort,
		Password: "", // no password set
		DB:       0,  // use default DB
	})

	// clean
	codeSlice, err := rdb.SMembers(ctx, "class").Result()
	if err != nil {
		logger.Error("Read redis set \"class\" fail", "error", err)
	}
	for _, code := range codeSlice {
		if err := rdb.Del(ctx, code).Err(); err != nil {
			logger.Errorf("Delete redis key \"%v\" fail", code, "error", err)
		}
	}
	rdb.Del(ctx, "class").Result()

	// get class code list
	data := getClassCode()
	for _, code := range data {
		if err = rdb.SAdd(ctx, "class", code.Code).Err(); err != nil {
			logger.Errorf("SAdd \"%v\" fail", code.Code, "error", err)
		}
	}

	// get free class
	for i := 0; i < 14; i++ {
		codeSlice, err := getFreeClassInfo(fmt.Sprint(i))
		if err != nil {
			logger.Panic("Fail getting free class info", "error", err)
		}
		for _, freeClass := range codeSlice {
			if err = rdb.SAdd(ctx, freeClass.Code, i).Err(); err != nil {
				logger.Errorf("SAdd set \"%v\" with \"%v\" fail", freeClass.Code, i, "error", err)

			}
		}
	}
}

// 第`classNo`节课的空闲教室
func getFreeClassInfo(classNo string) ([]ClassCode, error) {
	requestUrl := "https://app.upc.edu.cn/freeclass/wap/default/search-all"
	dateStr := time.Now().Format("2006-01-02")
	payload := url.Values{"xq": {"青岛校区"}, "date": {dateStr}, "lh": {"全部"}, "jc[]": {classNo}}
	resp, err := http.PostForm(requestUrl, payload)
	if err != nil {
		logger.Errorf("Fail getting free class info", "error", err, "requestUrl", requestUrl, "payload", payload)
	}
	defer resp.Body.Close()
	var respData ClassJsResp
	err = json.NewDecoder(resp.Body).Decode(&respData)
	if err != nil {
		logger.Errorf("Decode json fail", "error", err)

	}
	if respData.E == 0 {
		var result []ClassCode
		for _, v := range respData.D.Js {
			result = append(result, v...)
		}
		return result, nil
	}
	return nil, errors.New(respData.M)

}

func getClassCode() []ClassCode {
	requestUrl := "https://app.upc.edu.cn/freeclass/wap/default/search-all"
	resp, err := http.Get(requestUrl)
	if err != nil {
		logger.Errorf("Fail getting free class code info", "error", err, "requestUrl", requestUrl)
	}
	defer resp.Body.Close()
	var respData ClassAllResp
	err = json.NewDecoder(resp.Body).Decode(&respData)
	if err != nil {
		logger.Errorf("Decode json fail", "error", err)
	}
	var codeSlice []ClassCode
	if respData.E == 0 {
		for _, v := range respData.D.All {
			codeSlice = append(codeSlice, v...)
		}
	}
	return codeSlice
}
