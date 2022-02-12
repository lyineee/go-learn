package log_test

import (
	"errors"
	"os"
	"testing"

	"github.com/lyineee/go-learn/utils/log"
)

func TestPackageInfo(t *testing.T) {
	log.Info("test", log.String("test", "sdfsdf"))
	log.Default().Error("test err", log.Error(errors.New("sdfs")))
}

func TestDedicateConsoleError(t *testing.T) {
	l := log.NewLogger(log.NewConsoleCore(os.Stdout), log.InfoLevel)
	l.Error("test err", log.Error(errors.New("a err")))
}

func TestDedicateJsonError(t *testing.T) {
	l := log.NewLogger(log.NewJsonCore(os.Stdout), log.InfoLevel)
	l.Error("test err", log.Error(errors.New("a err")))
}
