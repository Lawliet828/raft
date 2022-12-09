package log

import (
	"io"
	"log"
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

type Config struct {
	Path     string
	FileName string
	Level    string
	MaxSize  int
}

var (
	logger      *zap.Logger
	sugarLogger *zap.SugaredLogger
)

func levelToZap(level string) zapcore.Level {
	switch level {
	case "debug":
		return zapcore.DebugLevel
	case "info":
		return zapcore.InfoLevel
	case "warn":
		return zapcore.WarnLevel
	case "error":
		return zapcore.ErrorLevel
	case "dpanic":
		return zapcore.DPanicLevel
	case "panic":
		return zapcore.PanicLevel
	case "fatal":
		return zapcore.FatalLevel
	default:
		return zapcore.InfoLevel
	}
}

func Init(conf Config) {
	writeSyncer := getLogWriter(conf)
	encoder := getEncoder()
	core := zapcore.NewCore(encoder, writeSyncer, levelToZap(conf.Level))

	logger = zap.New(core, zap.AddCaller(), zap.AddCallerSkip(1))
	sugarLogger = logger.Sugar()
}

func getEncoder() zapcore.Encoder {
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.EncodeTime = zapcore.TimeEncoderOfLayout("2006-01-02 15:04:05.000")
	encoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder
	return zapcore.NewConsoleEncoder(encoderConfig)
}

func getLogWriter(conf Config) zapcore.WriteSyncer {
	lumberJackLogger := &lumberjack.Logger{
		Filename:   conf.Path + "test.log",
		MaxSize:    conf.MaxSize,
		MaxBackups: 3,
		MaxAge:     30,
		Compress:   true,
	}
	ws := io.MultiWriter(os.Stdout, lumberJackLogger)
	return zapcore.AddSync(ws)
}

func GetStdLog() *log.Logger {
	return zap.NewStdLog(logger)
}

func Debug(args ...interface{}) {
	sugarLogger.Debug(args...)
	defer sugarLogger.Sync()
}

func Debugf(template string, args ...interface{}) {
	sugarLogger.Debugf(template, args...)
	defer sugarLogger.Sync()
}

func Info(args ...interface{}) {
	sugarLogger.Info(args...)
	defer sugarLogger.Sync()
}

func Infof(template string, args ...interface{}) {
	sugarLogger.Infof(template, args...)
	defer sugarLogger.Sync()
}

func Warn(args ...interface{}) {
	sugarLogger.Warn(args...)
	defer sugarLogger.Sync()
}

func Warnf(template string, args ...interface{}) {
	sugarLogger.Warnf(template, args...)
	defer sugarLogger.Sync()
}

func Error(args ...interface{}) {
	sugarLogger.Error(args...)
	defer sugarLogger.Sync()
}

func Errorf(template string, args ...interface{}) {
	sugarLogger.Errorf(template, args...)
	defer sugarLogger.Sync()
}

func DPanic(args ...interface{}) {
	sugarLogger.DPanic(args...)
	defer sugarLogger.Sync()
}

func DPanicf(template string, args ...interface{}) {
	sugarLogger.DPanicf(template, args...)
	defer sugarLogger.Sync()
}

func Panic(args ...interface{}) {
	sugarLogger.Panic(args...)
	defer sugarLogger.Sync()
}

func Panicf(template string, args ...interface{}) {
	sugarLogger.Panicf(template, args...)
	defer sugarLogger.Sync()
}
