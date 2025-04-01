// SPDX-FileCopyrightText: 2025 Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package log

import (
	"fmt"
	"io"
	"os"
	"regexp"
	"runtime"
	"strings"

	logLib "github.com/sirupsen/logrus"
	"r53restapi.com/pkg/buildflags"

	validation "github.com/go-ozzo/ozzo-validation/v4"
)

const (
	CALLER_OFFSET    = 12
	LOG_DIR_PERMS    = 0744
	LOG_FILE_PERMS   = 0644
	VALID_FILE_REGEX = `^[a-zA-Z0-9_-]+(\.log|\.txt)$`
	VALID_DIR_REGEX  = `^\/$|(\/[a-zA-Z_0-9-]+\/)+$`
)

var (
	allowedLevels      = []string{"info", "warning", "error"}
	allowedDebugLevels = []string{"debug", "info", "warning", "error"}
)

type Config struct {
	Directory  string
	File       string
	Level      string
	StdOutOnly bool
}

/*
Init initialises the logging package based on the Config struct it is passed.
*/
func Init(cfg Config) error {
	directoryPermissions := os.FileMode(LOG_DIR_PERMS)
	filePermissions := os.FileMode(LOG_FILE_PERMS)

	logLib.SetFormatter(DefaultFormat)
	logLib.SetReportCaller(true)

	if err := cfg.validate(); err != nil {
		Errorf("Log config validation error: %v", err)
		return err
	}

	if cfg.StdOutOnly  {
		Infof("Logging output is set to Stdout")
		logLib.SetOutput(os.Stdout)
	} else if cfg.File != "" {
		Infof("Setting log directory: %s", cfg.Directory)
		err := os.MkdirAll(cfg.Directory, directoryPermissions)
		if err != nil {
			Errorf("Error setting log directory: %v", err)
			return err
		}

		Infof("Setting log file: %s", cfg.File)
		fp, err := os.OpenFile(cfg.Directory+cfg.File, os.O_WRONLY|os.O_CREATE|os.O_APPEND, filePermissions)
		if err != nil {
			Errorf("Error setting log file: %v", err)
			return err
		}
		logLib.SetOutput(io.MultiWriter(fp, os.Stdout))
	}

	if cfg.Level != "" {
		Infof("Setting log level: %s", cfg.Level)
		level, err := logLib.ParseLevel(cfg.Level)
		if err != nil {
			Errorf("Error setting log level: %v", err)
			return err
		}
		logLib.SetLevel(level)

		if cfg.Level == "debug" && buildflags.DEBUG {
			Infof("Switching to debug log format")
			logLib.SetFormatter(DebugFormat)
		}
	}

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
	var logLevels []interface{}

	if buildflags.DEBUG {
		logLevels = make([]interface{}, len(allowedDebugLevels))
		for i, logLevel := range allowedDebugLevels {
			logLevels[i] = logLevel
		}
	} else {
		logLevels = make([]interface{}, len(allowedLevels))
		for i, logLevel := range allowedLevels {
			logLevels[i] = logLevel
		}
	}

	return validation.ValidateStruct(&c,
		validation.Field(
			&c.Directory,
			validation.Match(regexp.MustCompile(VALID_DIR_REGEX)).Error("must be a valid path"),
		),
		validation.Field(
			&c.File,
			validation.Match(regexp.MustCompile(VALID_FILE_REGEX)).Error("must be a valid filename"),
		),
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
