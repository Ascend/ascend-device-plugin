/*
* Copyright(C) 2020. Huawei Technologies Co.,Ltd. All rights reserved.
*
* Licensed under the Apache License, Version 2.0 (the "License");
* you may not use this file except in compliance with the License.
* You may obtain a copy of the License at
*
* http://www.apache.org/licenses/LICENSE-2.0
*
* Unless required by applicable law or agreed to in writing, software
* distributed under the License is distributed on an "AS IS" BASIS,
* WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
* See the License for the specific language governing permissions and
* limitations under the License.
 */

package huawei

import (
	"fmt"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
	"os"
	"path/filepath"
	"regexp"
	"time"
)

var logger *zap.Logger

// NewLogger to create logger
func NewLogger(loggerPath string) error {
	if !validate(loggerPath) {
		return fmt.Errorf("log path is error")
	}
	logger = ConfigLog(loggerPath)
	error := os.Chmod(loggerPath, logChmod)
	if error != nil && logger != nil {
		logger.Error("logger is error", zap.Error(error))
		return error
	}
	return nil
}

// GetLogger to get Logger
func GetLogger() *zap.Logger {
	return logger
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

func validate(path string) bool {
	relpath, err := filepath.Abs(path)
	if err != nil {
		fmt.Println("It's error when converted to an absolute path.")
		return false
	}
	pattern := `^/*`
	reg := regexp.MustCompile(pattern)
	return reg.MatchString(relpath)
}
