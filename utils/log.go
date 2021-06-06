package utils

import (
	"fmt"
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func Test() {
	fmt.Println("From package util")
}

func GetLogger() *zap.Logger {
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	consoleEncoder := zapcore.NewConsoleEncoder(encoderConfig)

	topicErrors := zapcore.AddSync(os.Stdout)
	core := zapcore.NewCore(consoleEncoder, topicErrors, zapcore.InfoLevel)

	log := zap.New(core, zap.AddCaller())
	return log
}
