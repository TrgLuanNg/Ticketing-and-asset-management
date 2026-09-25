package logger

import (
	"log"
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var Log *zap.SugaredLogger

func InitLogger(mode string) {
	var config zap.Config
	if mode == "release" {
		config = zap.NewProductionConfig()
	} else {
		config = zap.NewDevelopmentConfig()
		config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	}

	config.EncoderConfig.TimeKey = "timestamp"
	config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder

	zapLogger, err := config.Build()
	if err != nil {
		log.Printf("Failed to initialize zap logger: %v, falling back to standard logger", err)
		Log = zap.NewExample().Sugar()
		return
	}

	Log = zapLogger.Sugar()
}

func Sync() {
	if Log != nil {
		_ = Log.Sync()
	}
}

func Info(args ...interface{}) {
	if Log != nil {
		Log.Info(args...)
	} else {
		log.Println(args...)
	}
}

func Infof(template string, args ...interface{}) {
	if Log != nil {
		Log.Infof(template, args...)
	} else {
		log.Printf(template, args...)
	}
}

func Error(args ...interface{}) {
	if Log != nil {
		Log.Error(args...)
	} else {
		log.Println(args...)
	}
}

func Errorf(template string, args ...interface{}) {
	if Log != nil {
		Log.Errorf(template, args...)
	} else {
		log.Printf(template, args...)
	}
}

func Fatal(args ...interface{}) {
	if Log != nil {
		Log.Fatal(args...)
	} else {
		log.Fatalln(args...)
		os.Exit(1)
	}
}

func Fatalf(template string, args ...interface{}) {
	if Log != nil {
		Log.Fatalf(template, args...)
	} else {
		log.Fatalf(template, args...)
		os.Exit(1)
	}
}

func Warn(args ...interface{}) {
	if Log != nil {
		Log.Warn(args...)
	} else {
		log.Println(args...)
	}
}

func Warnf(template string, args ...interface{}) {
	if Log != nil {
		Log.Warnf(template, args...)
	} else {
		log.Printf(template, args...)
	}
}

func Debug(args ...interface{}) {
	if Log != nil {
		Log.Debug(args...)
	} else {
		log.Println(args...)
	}
}

func Debugf(template string, args ...interface{}) {
	if Log != nil {
		Log.Debugf(template, args...)
	} else {
		log.Printf(template, args...)
	}
}
