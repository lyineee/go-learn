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

func main() {
	var ctx = context.Background()
	rdb := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "", // no password set
		DB:       0,  // use default DB
	})
	// clean
	codeSlice, err := rdb.SMembers(ctx, "class").Result()
	if err != nil {
		fmt.Println(err)
	}
	for _, code := range codeSlice {
		if err := rdb.Del(ctx, code).Err(); err != nil {
			fmt.Println(err)
		}
	}
	rdb.Del(ctx, "class").Result()

	// get class code list
	data := getClassCode()
	for _, code := range data {
		if err = rdb.SAdd(ctx, "class", code.Code).Err(); err != nil {
			fmt.Println(err)
		}
	}

	// get free class
	for i := 0; i < 14; i++ {
		codeSlice, err := getFreeClassInfo(fmt.Sprint(i))
		if err != nil {
			fmt.Println(err)
		}
		for _, freeClass := range codeSlice {
			if err = rdb.SAdd(ctx, freeClass.Code, i).Err(); err != nil {
				fmt.Println(err)
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
		fmt.Println("Error!", err)
	}
	defer resp.Body.Close()
	var respData ClassJsResp
	err = json.NewDecoder(resp.Body).Decode(&respData)
	if err != nil {
		fmt.Println("Error!", err)
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
	resp, err := http.Get("https://app.upc.edu.cn/freeclass/wap/default/search-all")
	if err != nil {
		fmt.Println("Error!", err)
	}
	defer resp.Body.Close()
	var respData ClassAllResp
	err = json.NewDecoder(resp.Body).Decode(&respData)
	if err != nil {
		fmt.Println("Error!", err)
	}
	var codeSlice []ClassCode
	if respData.E == 0 {
		for _, v := range respData.D.All {
			codeSlice = append(codeSlice, v...)
		}
	}
	return codeSlice
}
