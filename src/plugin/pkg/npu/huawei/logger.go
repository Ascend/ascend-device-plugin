/*
 * Copyright (c) Huawei Technologies Co., Ltd. 2019-2024. All rights reserved.
 * Description: logger.go
 * Create: 19-11-21 上午10:46
 */

package huawei

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
	"os"
	"time"
)

var logger *zap.Logger

func init() {
	logger = ConfigLog(LogPath)
	error := os.Chmod(LogPath, logChmod)
	if error != nil {
		logger.Error("logger is error", zap.Error(error))
	}
}

// NewEncoderConfig is used to config log file
func NewEncoderConfig() zapcore.EncoderConfig {
	return zapcore.EncoderConfig{
		// Keys can be anything except the empty string.
		TimeKey:        "T",
		LevelKey:       "L",
		NameKey:        "N",
		CallerKey:      "C",
		MessageKey:     "M",
		StacktraceKey:  "S",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.CapitalLevelEncoder,
		EncodeTime:     TimeEncoder,
		EncodeDuration: zapcore.StringDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}
}

// TimeEncoder is used to specify time
func TimeEncoder(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
	enc.AppendString(t.Format("2006-01-02 15:04:05.000"))
}

// ConfigLog is used to config zap log
func ConfigLog(logPath string) *zap.Logger {
	w := zapcore.AddSync(&lumberjack.Logger{
		Filename:   logPath,
		MaxSize:    fileMaxSize, // megabytes
		MaxBackups: maxBackups,
		MaxAge:     maxAge, // days
	})
	core := zapcore.NewCore(
		zapcore.NewConsoleEncoder(NewEncoderConfig()),
		zapcore.NewMultiWriteSyncer(zapcore.AddSync(os.Stdout),
			w),
		zap.InfoLevel,
	)
	return zap.New(core, zap.AddCaller())
}
