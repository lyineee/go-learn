package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"errors"

	"github.com/lyineee/go-learn/utils/log"
	"github.com/spf13/viper"
)

type statusResp struct {
	Data struct {
		RoomInfo struct {
			LiveStatus int `json:"live_status"`
		} `json:"room_info"`
	} `json:"data"`
}

var (
	crawlInterval int
	crawlTimeout  int
	roomId        int
	barkToken     string
	waitCount     int
)

func main() {
	initConfig()
	chanMsg := make(chan bool)
	chanErr := make(chan error)
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() { // live status goroutine
		log.Info("start crawling live status", log.Int("room id", roomId))
		for {
			ctx, cancel := context.WithTimeout(context.TODO(), time.Duration(crawlTimeout)*time.Second)
			status, err := getRoomStatus(ctx, roomId)
			if err != nil {
				chanErr <- err
				cancel()
				continue
			}
			chanMsg <- status
			cancel()
			time.Sleep(time.Duration(crawlInterval) * time.Second)
		}
	}()
	waitConuntDown := 0
	for {
		select {
		case status := <-chanMsg:
			if waitConuntDown == 0 && status {
				waitConuntDown = waitCount
			}
			if waitConuntDown > 0 && status {
				err := mkNotification()
				if err != nil {
					log.Default().Error("error when push to bark", log.Error(err))
					continue
				}
			} else {
				waitConuntDown--
			}
		case err := <-chanErr:
			log.Default().Error("error", log.Error(err))
		case <-sigs:
			log.Info("graceful shutdown")
			return
		}
	}
}

func initConfig() {
	viper.SetDefault("crawlInterval", 5)
	viper.SetDefault("crawlTimeout", 2)
	viper.SetDefault("roomId", 92613)
	viper.SetDefault("barkToken", "sdfsfd")
	viper.SetDefault("waitCount", 3)

	viper.SetConfigFile("/etc/live-notification.toml")

	crawlInterval = viper.GetInt("crawlInterval")
	crawlTimeout = viper.GetInt("crawlTimeout")
	roomId = viper.GetInt("roomId")
	barkToken = viper.GetString("barkToken")
	waitCount = viper.GetInt("waitCount")
	log.Info("init config", log.Any("config map", viper.AllSettings()))
}

func mkNotification() error {
	url := fmt.Sprintf("https://live.bilibili.com/%d", roomId)
	urlPara := fmt.Sprintf("%s/%s?url=%s", "ZBL", url, url)
	log.Info("debug", log.String("bark url parameter", urlPara))
	ctx, cancel := context.WithTimeout(context.TODO(), time.Duration(crawlTimeout)*time.Second)
	defer cancel()
	err := pushBark(ctx, barkToken, urlPara)
	if err != nil {
		return err
	}
	return nil
}

func getRoomStatus(ctx context.Context, roomId int) (status bool, err error) {
	resp, err := http.Get(fmt.Sprintf("https://api.live.bilibili.com/xlive/web-room/v1/index/getInfoByRoom?room_id=%d", roomId))
	if err != nil {
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return status, errors.New("http request fail")
	}
	statusResp := statusResp{}
	codec := json.NewDecoder(resp.Body)
	err = codec.Decode(&statusResp)
	if err != nil {
		return
	}
	status = statusResp.Data.RoomInfo.LiveStatus != 0
	return
}

func pushBark(ctx context.Context, token, urlPara string) error {
	resp, err := http.Get(fmt.Sprintf("https://api.day.app/%s/%s", token, urlPara))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		c, _ := ioutil.ReadAll(resp.Body)
		log.Default().Error("bark return with non 200", log.String("token", token), log.String("urlPara", urlPara), log.String("response content", string(c)))
		return errors.New("bark return with non 200")
	}
	return nil
}
