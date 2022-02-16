package main_test

import (
	"context"
	"encoding/json"
	"testing"

	loki "github.com/lyineee/go-learn/loki-redis"
)

func TestJsonMarshal(t *testing.T) {
	a := loki.Streams{
		[]loki.StreamItem{{
			[][2]string{{"13213213", "fdalkjljlkjlsf"}},
			loki.Label{Subject: "test-subject"},
		}},
	}
	b, err := json.Marshal(a)
	if err != nil {
		t.Error(err)
	}
	t.Log(string(b))
}

func TestPush(t *testing.T) {
	a := loki.Streams{
		[]loki.StreamItem{{
			[][2]string{{"13213213", "fdalkjljlkjlsf"}},
			loki.Label{Subject: "test-subject"},
		}},
	}
	loki.Push(context.Background(), "http://localhost:3100/loki/api/v1/push", &a)
}
