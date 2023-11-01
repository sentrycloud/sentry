package newlog

import (
	"gopkg.in/natefinch/lumberjack.v2"
	"log"
	"os"
	"path"
	"strings"
)

type LogLevel int

const (
	LevelError LogLevel = iota
	LevelWarn
	LevelInfo
	LevelDebug
)

var (
	logLevel = LevelInfo // default set to info level
)

type LogConfig struct {
	Path      string `json:"path"`
	Level     string `json:"level"`
	MaxSize   int    `json:"max_size"`
	MaxBackup int    `json:"max_backup"`
}

func setStrLogLevel(level string) {
	level = strings.ToLower(level)
	switch level {
	case "error":
		logLevel = LevelError
	case "warn":
		logLevel = LevelWarn
	case "info":
		logLevel = LevelInfo
	case "debug":
		logLevel = LevelDebug
	default:
		logLevel = LevelInfo
	}
}

func SetConfig(logConfig *LogConfig) {
	// lumberjack create path with mode 0744, so other user can not read the log, change path mode to 755
	dir := path.Dir(logConfig.Path)
	if len(dir) > 0 {
		os.MkdirAll(dir, 0755)
	}

	log.SetOutput(&lumberjack.Logger{
		Filename:   logConfig.Path,
		MaxSize:    logConfig.MaxSize,
		MaxBackups: logConfig.MaxBackup,
		LocalTime:  true,
	})

	setStrLogLevel(logConfig.Level)
}

func Fatal(format string, v ...interface{}) {
	log.Printf("[FATAL] "+format, v...)
	os.Exit(1)
}

func Error(format string, v ...interface{}) {
	if logLevel >= LevelError {
		log.Printf("[ERROR] "+format, v...)
	}
}

func Warn(format string, v ...interface{}) {
	if logLevel >= LevelWarn {
		log.Printf("[WARN] "+format, v...)
	}
}

func Info(format string, v ...interface{}) {
	if logLevel >= LevelInfo {
		log.Printf("[Info] "+format, v...)
	}
}

func Debug(format string, v ...interface{}) {
	if logLevel >= LevelDebug {
		log.Printf("[Debug] "+format, v...)
	}
}
