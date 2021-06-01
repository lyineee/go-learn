package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"
	"time"

	"github.com/goinggo/mapstructure"
)

type ClassCode struct {
	Code string `mapstructure:"code"`
	Name string `mapstructure:"name"`
}

type ClassResp struct {
	E int
	M string
	D interface{}
}

func main() {
	fmt.Println(time.Now().Format("2006-01-02"))
	data := getClassCode()
	fmt.Println(data)
}

func getFreeClassInfo(classNo string) {}

func getClassCode() []ClassCode {
	resp, err := http.Get("https://app.upc.edu.cn/freeclass/wap/default/search-all")
	if err != nil {
		fmt.Println(err)
	}
	defer resp.Body.Close()
	var data ClassResp
	err = json.NewDecoder(resp.Body).Decode(&data)
	if err != nil {
		fmt.Println("Error!", err)
	}
	var codeSlice []ClassCode
	if data.E == 0 {
		v := reflect.ValueOf(data.D)
		if v.Kind() == reflect.Map {
			for _, key := range v.MapKeys() {
				if key.Interface().(string) == "all" {
					locationMap := v.MapIndex(key).Interface().(map[string]interface{})
					for _, location := range locationMap {
						var temp ClassCode
						content := location.([]interface{})
						for _, d := range content {
							mapstructure.Decode(d, &temp)
							codeSlice = append(codeSlice, temp)
						}
					}
					break
				}
			}
		}
	}
	return codeSlice
}
