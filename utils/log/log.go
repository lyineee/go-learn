package log

import (
	"io"
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type Logger struct {
	l     *zap.Logger
	level Level
}
type Level = zapcore.Level
type Field = zap.Field

type SugarLogger = zap.SugaredLogger

//field type sugar
var (
	Any    = zap.Any
	String = zap.String
	Int    = zap.Int
	Bool   = zap.Bool
	Error  = zap.Error
	Skip   = zap.Skip
)

//log level
var (
	InfoLevel         = zap.InfoLevel   // 0
	WarnLevel   Level = zap.WarnLevel   // 1
	ErrorLevel  Level = zap.ErrorLevel  // 2
	DPanicLevel Level = zap.DPanicLevel // 3, used in development log
	// PanicLevel logs a message, then panics
	PanicLevel Level = zap.PanicLevel // 4
	// FatalLevel logs a message, then calls os.Exit(1).
	FatalLevel Level = zap.FatalLevel // 5
	DebugLevel Level = zap.DebugLevel // -1
)

var std = NewLogger(NewJsonCore(os.Stdout), InfoLevel)

//expose in package level, no error level
var (
	Info  = std.Info
	Panic = std.Panic
	Fatal = std.Fatal
)

func (l *Logger) Info(msg string, fields ...Field) {
	l.l.Info(msg, fields...)
}

func (l *Logger) Error(msg string, fields ...Field) {
	l.l.Error(msg, fields...)
}

func (l *Logger) Panic(msg string, fields ...Field) {
	l.l.Panic(msg, fields...)
}

func (l *Logger) Fatal(msg string, fields ...Field) {
	l.l.Fatal(msg, fields...)
}

func (l *Logger) Sugar() *SugarLogger {
	return (*SugarLogger)(l.l.Sugar())
}

func (l *Logger) Sync() error {
	return l.l.Sync()
}

func Sync() error {
	if std != nil {
		return std.Sync()
	}
	return nil
}

func NewLogger(core zapcore.Core, level Level) *Logger {
	l := zap.New(core, zap.AddCaller(), zap.AddCallerSkip(1))
	logger := Logger{
		l:     l,
		level: level,
	}
	return &logger
}

func NewConsoleCore(w io.Writer) zapcore.Core {
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	consoleEncoder := zapcore.NewConsoleEncoder(encoderConfig)

	topicErrors := zapcore.AddSync(w)
	core := zapcore.NewCore(consoleEncoder, topicErrors, zapcore.InfoLevel)
	return core
}

func NewJsonCore(w io.Writer) zapcore.Core {
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	consoleEncoder := zapcore.NewJSONEncoder(encoderConfig)

	topicErrors := zapcore.AddSync(w)
	core := zapcore.NewCore(consoleEncoder, topicErrors, zapcore.InfoLevel)
	return core
}

func Default() *Logger { return std }

func ReplaceDefault(l *Logger) { std = l }
