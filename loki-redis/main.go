package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"time"

	rstream "github.com/lyineee/go-learn/redis-stream"
	"github.com/lyineee/go-learn/utils/log"
	_ "github.com/lyineee/go-learn/utils/remote"
	"github.com/spf13/viper"
)

type Label struct {
	Subject string `json:"subject"`
}
type Streams struct {
	Streams []StreamItem `json:"streams"`
}

type StreamItem struct {
	Values [][2]string `json:"values"`
	Stream Label       `json:"stream"`
}

type Ts struct {
	Ts string `json:"ts"`
}

var logger = log.NewLogger(log.NewJsonCore(os.Stdout), log.InfoLevel)

func main() {
	// get envirment
	viper.AutomaticEnv()

	// init etcd config
	viper.SetDefault("etcd", "etcd:2379")
	viper.SetDefault("etcd_config_path", "/config/loki-redis.toml")

	//database
	viper.SetDefault("database.redis", "redis:6379")

	//stream
	viper.SetDefault("stream.stream", "stream.log")
	viper.SetDefault("stream.group", "stream.log.worker")

	//loki
	viper.SetDefault("loki.address", "http://localhost:3100")

	viper.AddRemoteProvider("etcd", viper.GetString("etcd"), viper.GetString("etcd_config_path"))
	viper.SetConfigType("toml")
	err := viper.ReadRemoteConfig()
	if err != nil {
		log.Panic("cannot connect to etcd config center", log.String("etcd", viper.GetString("etcd")), log.String("etcd_path", viper.GetString("etcd_config_path")), log.Error(err))
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	group, err := rstream.NewGroup(ctx, viper.GetString("database.redis"), viper.GetString("database.redis_password"), viper.GetString("stream.group"), viper.GetString("stream.stream"))
	if err != nil {
		logger.Fatal("err", log.Error(err))
	}

	for i := range group.Subscribe() {
		if i.Error != nil {
			logger.Error("err", log.Error(i.Error))
			continue
		}
		for key := range i.Values {
			ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
			item := StreamItem{}
			ts := Ts{}
			item.Stream.Subject = key //subject key
			err := json.Unmarshal([]byte(i.Values[key].(string)), &ts)
			if err != nil {
				logger.Error("unmarshal json error", log.String("row_json", i.Values[key].(string)), log.Error(err))
				continue
			}
			t, err := time.Parse("2006-01-02T15:04:05.999-0700", ts.Ts)
			if err != nil {
				logger.Error("parse time error", log.Error(err))
				continue
			}
			line := [2]string{strconv.FormatInt(t.UnixNano(), 10), i.Values[key].(string)}
			item.Values = make([][2]string, 1) //line init
			item.Values[0] = line
			result := Streams{[]StreamItem{item}}
			js, err := json.Marshal(result)
			if err != nil {
				logger.Error("marshal payload error", log.Error(err))
				continue
			}
			logger.Info("result", log.Any("js", string(js)))
			loki := viper.GetString("loki.address") + "/loki/api/v1/push"
			err = Push(ctx, loki, &result)
			if err != nil {
				logger.Error("error when push to loki", log.String("loki_url", loki), log.Error(err))
				continue
			}
			group.Ack(ctx, i.ID)
			cancel()
		}
	}
}

//push line to loki instance
func Push(ctx context.Context, url string, streams *Streams) error {
	payload, err := json.Marshal(streams)
	if err != nil {
		return err
	}
	resp, err := http.Post(url, "application/json", bytes.NewReader(payload))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	content, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	if string(content) != "" {
		logger.Info("Push loki fail", log.Any("payload", string(payload)), log.String("response_content", string(content)))
		return errors.New("Push loki fail")
	}
	return nil
}
