// Copyright (C) 2025 Intel Corporation
// SPDX-FileCopyrightText: 2025 Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package log

import (
	"fmt"
	"os"
	"runtime"
	"strings"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	logLib "github.com/sirupsen/logrus"
)

const CALLER_OFFSET = 12

var allowedLevels = []string{"debug", "info", "warning", "error"}

type Config struct {
	Level string
}

/*
Init initialises the logging package based on the Config struct it is passed.
*/
func Init(cfg Config) error {

	logLib.SetFormatter(DefaultFormat)
	logLib.SetReportCaller(true)

	if err := cfg.validate(); err != nil {
		Errorf("Log config validation error: %v", err)
		return err
	}

	if cfg.Level != "" {
		Infof("Setting log level: %s", cfg.Level)
		level, err := logLib.ParseLevel(cfg.Level)
		if err != nil {
			Errorf("Error setting log level: %v", err)
			return err
		}
		logLib.SetLevel(level)

		if cfg.Level == "debug" {
			Infof("Switching to debug log format")
			logLib.SetFormatter(DebugFormat)
		}
	}

	logLib.SetOutput(os.Stdout)

	return nil
}

/*
DefaultFormat is the default formatter for our logs.
*/
var DefaultFormat = &logLib.TextFormatter{
	FullTimestamp:   true,
	TimestampFormat: "2006-01-02 15:04:05",
	ForceColors:     true,
	CallerPrettyfier: func(frame *runtime.Frame) (string, string) {
		functionName, _ := callerInfo()
		return functionName, ""
	},
}

/*
DebugFormat is the debug formatter for our logs, it prints additional data to aid with debugging.
*/
var DebugFormat = &logLib.TextFormatter{
	FullTimestamp:    true,
	TimestampFormat:  "2006-01-02 15:04:05",
	ForceColors:      true,
	CallerPrettyfier: func(frame *runtime.Frame) (string, string) { return callerInfo() },
}

/*
TestFormat is the formatter used in tests.
*/
var TestFormat = &logLib.TextFormatter{
	DisableColors: true,
	FullTimestamp: true,
}

/*
Debugf wraps Debugf.
*/
func Debugf(msg string, args ...interface{}) {
	logLib.Debugf(msg, args...)
}

/*
Infof wraps Infof.
*/
func Infof(msg string, args ...interface{}) {
	logLib.Infof(msg, args...)
}

/*
Warnf wraps Warnf.
*/
func Warnf(msg string, args ...interface{}) {
	logLib.Warnf(msg, args...)
}

/*
Errorf wraps Errorf.
*/
func Errorf(msg string, args ...interface{}) {
	logLib.Errorf(msg, args...)
}

/*
validate validates the contents of the Config struct.
*/
func (c Config) validate() error {
	logLevels := make([]interface{}, len(allowedLevels))

	for i, logLevel := range allowedLevels {
		logLevels[i] = logLevel
	}

	return validation.ValidateStruct(&c,
		validation.Field(
			&c.Level,
			validation.In(logLevels...).Error("must be one of "+fmt.Sprintf("%v", logLevels)),
		),
	)
}

/*
callerInfo fixes the caller offset resulting from the logLib functions being called withing this package.
Keeps file, line and function details correct in debug logs .
*/
func callerInfo() (string, string) {
	var funcName, fileName string
	pc, file, line, ok := runtime.Caller(CALLER_OFFSET)
	if !ok {
		return "[?]", " [?:?]"
	} else {
		slash := strings.LastIndex(file, "/")
		if slash >= 0 {
			fileName = file[slash+1:]
		}

		details := runtime.FuncForPC(pc)
		if details != nil {
			s := strings.Split(details.Name(), ".")
			funcName = s[len(s)-1]
		}
	}
	return fmt.Sprintf("[%s]", funcName), fmt.Sprintf(" [%s:%d]", fileName, line)
}
