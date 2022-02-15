package remote

import (
	"bytes"
	"context"
	"io"
	"time"

	"github.com/spf13/viper"
	"go.etcd.io/etcd/clientv3"
)

type Config struct{}

func (c *Config) Get(rp viper.RemoteProvider) (io.Reader, error) {
	client, err := c.newClient(rp)
	if err != nil {
		return nil, err
	}
	defer client.Close()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	resp, err := client.Get(ctx, rp.Path())
	cancel()
	if err != nil {
		return nil, err
	}
	return bytes.NewReader(resp.Kvs[0].Value), nil
}

func (c *Config) Watch(rp viper.RemoteProvider) (io.Reader, error) {
	client, err := c.newClient(rp)
	if err != nil {
		return nil, err
	}
	defer client.Close()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	resp, err := client.Get(ctx, rp.Path())
	cancel()
	if err != nil {
		return nil, err
	}
	return bytes.NewReader(resp.Kvs[0].Value), nil
}

func (c *Config) WatchChannel(rp viper.RemoteProvider) (<-chan *viper.RemoteResponse, chan bool) {

	rr := make(chan *viper.RemoteResponse)
	stop := make(chan bool)

	go func() {
		for {
			client, err := c.newClient(rp)

			if err != nil {
				time.Sleep(time.Duration(time.Second))
				continue
			}

			defer client.Close()

			ch := client.Watch(context.Background(), rp.Path())

			select {
			case <-stop:
				return
			case res := <-ch:
				for _, event := range res.Events {
					rr <- &viper.RemoteResponse{
						Value: event.Kv.Value,
					}
				}
			}
		}
	}()

	return rr, stop
}

func (c *Config) newClient(rp viper.RemoteProvider) (*clientv3.Client, error) {
	client, err := clientv3.New(clientv3.Config{
		Endpoints: []string{rp.Endpoint()},
	})

	if err != nil {
		return nil, err
	}
	return client, nil
}
