package remote_test

import (
	"testing"

	_ "github.com/lyineee/go-learn/utils/remote"
	"github.com/spf13/viper"
)

func TestRemote(t *testing.T) {
	viper.AddRemoteProvider("etcd", "localhost:2379", "/config/history-publisher.toml")
	viper.SetConfigType("toml")
	err := viper.ReadRemoteConfig()
	if err != nil {
		t.Error("fail", err)
	}
}
