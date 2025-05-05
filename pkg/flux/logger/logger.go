package logger

import (
	"fmt"
	"io"
	"os"
	"strings"
	"sync"
	"time"
)

type Level int

const (
	LevelDebug Level = iota

	LevelInfo

	LevelWarn

	LevelError

	LevelFatal
)

// representation of the log level
func (l Level) String() string {
	switch l {
	case LevelDebug:
		return "DEBUG"
	case LevelInfo:
		return "INFO"
	case LevelWarn:
		return "WARN"
	case LevelError:
		return "ERROR"
	case LevelFatal:
		return "FATAL"
	default:
		return "UNKNOWN"
	}
}

func ParseLevel(level string) Level {
	switch strings.ToUpper(level) {
	case "DEBUG":
		return LevelDebug
	case "INFO":
		return LevelInfo
	case "WARN":
		return LevelWarn
	case "ERROR":
		return LevelError
	case "FATAL":
		return LevelFatal
	default:
		return LevelInfo
	}
}

type Fields map[string]interface{}

type Logger struct {
	level      Level
	writer     io.Writer
	fields     Fields
	timeFormat string
	mu         sync.Mutex
	colorized  bool
}

// Config
type Config struct {
	Level      Level
	Writer     io.Writer
	TimeFormat string
	Colorized  bool
}

func DefaultConfig() Config {
	return Config{
		Level:      LevelInfo,
		Writer:     os.Stdout,
		TimeFormat: "2006-01-02 15:04:05",
		Colorized:  true,
	}
}

// logger instance
func New(config Config) *Logger {
	if config.Writer == nil {
		config.Writer = os.Stdout
	}
	if config.TimeFormat == "" {
		config.TimeFormat = "2006-01-02 15:04:05"
	}
	return &Logger{
		level:      config.Level,
		writer:     config.Writer,
		fields:     make(Fields),
		timeFormat: config.TimeFormat,
		colorized:  config.Colorized,
	}
}

// default configuration
func DefaultLogger() *Logger {
	config := DefaultConfig()
	return New(config)
}

func (l *Logger) WithLevel(level Level) *Logger {
	newFields := make(Fields)
	for k, v := range l.fields {
		newFields[k] = v
	}

	return &Logger{
		level:      level,
		writer:     l.writer,
		fields:     newFields,
		timeFormat: l.timeFormat,
		colorized:  l.colorized,
	}
}

func (l *Logger) WithField(key string, value interface{}) *Logger {
	newFields := make(Fields)
	for k, v := range l.fields {
		newFields[k] = v
	}
	newFields[key] = value

	return &Logger{
		level:      l.level,
		writer:     l.writer,
		fields:     newFields,
		timeFormat: l.timeFormat,
		colorized:  l.colorized,
	}
}

func (l *Logger) WithFields(fields Fields) *Logger {
	newFields := make(Fields)
	for k, v := range l.fields {
		newFields[k] = v
	}
	for k, v := range fields {
		newFields[k] = v
	}

	return &Logger{
		level:      l.level,
		writer:     l.writer,
		fields:     newFields,
		timeFormat: l.timeFormat,
		colorized:  l.colorized,
	}
}

func (l *Logger) Debug(message string, args ...interface{}) {
	l.log(LevelDebug, message, args...)
}

func (l *Logger) Info(message string, args ...interface{}) {
	l.log(LevelInfo, message, args...)
}

func (l *Logger) Warn(message string, args ...interface{}) {
	l.log(LevelWarn, message, args...)
}

func (l *Logger) Error(message string, args ...interface{}) {
	l.log(LevelError, message, args...)
}

func (l *Logger) Fatal(message string, args ...interface{}) {
	l.log(LevelFatal, message, args...)
	os.Exit(1)
}

func (l *Logger) log(level Level, message string, args ...interface{}) {
	if level < l.level {
		return
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	if len(args) > 0 {
		message = fmt.Sprintf(message, args...)
	}

	timestamp := time.Now().Format(l.timeFormat)
	levelStr := level.String()

	var coloredLevel string
	if l.colorized {
		coloredLevel = l.colorize(levelStr, level)
	} else {
		coloredLevel = levelStr
	}

	line := fmt.Sprintf("[%s] [%s] %s", timestamp, coloredLevel, message)

	if len(l.fields) > 0 {
		fieldStr := ""
		for k, v := range l.fields {
			fieldStr += fmt.Sprintf(" %s=%v", k, v)
		}
		line += fieldStr
	}

	line += "\n"

	fmt.Fprint(l.writer, line)
}

func (l *Logger) colorize(level string, logLevel Level) string {
	var colorCode string

	switch logLevel {
	case LevelDebug:
		colorCode = "\033[37m"
	case LevelInfo:
		colorCode = "\033[32m"
	case LevelWarn:
		colorCode = "\033[33m"
	case LevelError:
		colorCode = "\033[31m"
	case LevelFatal:
		colorCode = "\033[35m"
	default:
		colorCode = "\033[0m"
	}

	resetCode := "\033[0m"
	return fmt.Sprintf("%s%s%s", colorCode, level, resetCode)
}

// Global logger instance
var (
	global     *Logger
	globalOnce sync.Once
)

func Global() *Logger {
	globalOnce.Do(func() {
		global = DefaultLogger()
	})
	return global
}

func SetGlobal(logger *Logger) {
	global = logger
}

func LogDebug(message string, args ...interface{}) {
	Global().Debug(message, args...)
}

func LogInfo(message string, args ...interface{}) {
	Global().Info(message, args...)
}

func LogWarn(message string, args ...interface{}) {
	Global().Warn(message, args...)
}

func LogError(message string, args ...interface{}) {
	Global().Error(message, args...)
}

func LogFatal(message string, args ...interface{}) {
	Global().Fatal(message, args...)
}
