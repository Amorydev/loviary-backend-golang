package logger

import (
	"encoding/json"
	"log"
	"os"
	"time"
)

// Level represents log level
type Level int

const (
	LevelDebug Level = iota
	LevelInfo
	LevelWarn
	LevelError
)

// Logger provides structured logging
type Logger struct {
	level Level
}

// New creates a new logger
func New(levelStr, format string) *Logger {
	var lvl Level
	switch levelStr {
	case "debug":
		lvl = LevelDebug
	case "info":
		lvl = LevelInfo
	case "warn":
		lvl = LevelWarn
	case "error":
		lvl = LevelError
	default:
		lvl = LevelInfo
	}
	return &Logger{level: lvl}
}

// log writes a structured log entry
func (l *Logger) log(level Level, msg string, fields map[string]interface{}) {
	if level < l.level {
		return
	}

	logData := map[string]interface{}{
		"level":   level.String(),
		"message": msg,
		"time":    time.Now().UTC().Format(time.RFC3339),
	}
	for k, v := range fields {
		logData[k] = v
	}

	if jsonData, err := json.Marshal(logData); err == nil {
		log.Println(string(jsonData))
	} else {
		log.Printf("[%s] %s %v", level.String(), msg, fields)
	}
}

// Debug logs debug message
func (l *Logger) Debug(msg string, fields ...map[string]interface{}) {
	var f map[string]interface{}
	if len(fields) > 0 {
		f = fields[0]
	}
	l.log(LevelDebug, msg, f)
}

// Info logs info message
func (l *Logger) Info(msg string, fields ...map[string]interface{}) {
	var f map[string]interface{}
	if len(fields) > 0 {
		f = fields[0]
	}
	l.log(LevelInfo, msg, f)
}

// Warn logs warning message
func (l *Logger) Warn(msg string, fields ...map[string]interface{}) {
	var f map[string]interface{}
	if len(fields) > 0 {
		f = fields[0]
	}
	l.log(LevelWarn, msg, f)
}

// Error logs error message
func (l *Logger) Error(msg string, err error, fields ...map[string]interface{}) {
	var f map[string]interface{}
	if len(fields) > 0 {
		f = fields[0]
	}
	if f == nil {
		f = make(map[string]interface{})
	}
	if err != nil {
		f["error"] = err.Error()
	}
	l.log(LevelError, msg, f)
}

// Fatal logs error and exits
func (l *Logger) Fatal(msg string, err error) {
	l.Error(msg, err, nil)
	os.Exit(1)
}

func (l Level) String() string {
	switch l {
	case LevelDebug:
		return "debug"
	case LevelInfo:
		return "info"
	case LevelWarn:
		return "warn"
	case LevelError:
		return "error"
	default:
		return "unknown"
	}
}

// Close is a no-op but kept for interface compatibility
func (l *Logger) Close() {}
