package logger

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/swarch-2f-rootly/rootly-apigateway/internal/core/ports"
)

// Logger implements structured logging
type Logger struct {
	level      LogLevel
	format     LogFormat
	output     *log.Logger
	component  string
}

// LogLevel represents logging levels
type LogLevel int

const (
	DebugLevel LogLevel = iota
	InfoLevel
	WarnLevel
	ErrorLevel
)

// LogFormat represents logging formats
type LogFormat int

const (
	JSONFormat LogFormat = iota
	TextFormat
)

// NewLogger creates a new logger instance
func NewLogger(level string, format string, component string) ports.Logger {
	logger := &Logger{
		level:     parseLogLevel(level),
		format:    parseLogFormat(format),
		output:    log.New(os.Stdout, "", 0),
		component: component,
	}

	return logger
}

// Debug logs debug messages
func (l *Logger) Debug(msg string, fields map[string]interface{}) {
	if l.level > DebugLevel {
		return
	}
	l.log(DebugLevel, msg, nil, fields)
}

// Info logs info messages
func (l *Logger) Info(msg string, fields map[string]interface{}) {
	if l.level > InfoLevel {
		return
	}
	l.log(InfoLevel, msg, nil, fields)
}

// Warn logs warning messages
func (l *Logger) Warn(msg string, fields map[string]interface{}) {
	if l.level > WarnLevel {
		return
	}
	l.log(WarnLevel, msg, nil, fields)
}

// Error logs error messages
func (l *Logger) Error(msg string, err error, fields map[string]interface{}) {
	l.log(ErrorLevel, msg, err, fields)
}

// log performs the actual logging
func (l *Logger) log(level LogLevel, msg string, err error, fields map[string]interface{}) {
	logEntry := LogEntry{
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Level:     l.levelToString(level),
		Message:   msg,
		Component: l.component,
		Fields:    fields,
	}

	if err != nil {
		logEntry.Error = err.Error()
	}

	var output string
	switch l.format {
	case JSONFormat:
		if jsonBytes, err := json.Marshal(logEntry); err == nil {
			output = string(jsonBytes)
		} else {
			output = l.formatAsText(logEntry)
		}
	case TextFormat:
		output = l.formatAsText(logEntry)
	default:
		output = l.formatAsText(logEntry)
	}

	l.output.Println(output)
}

// formatAsText formats log entry as text
func (l *Logger) formatAsText(entry LogEntry) string {
	result := fmt.Sprintf("[%s] %s %s", entry.Timestamp, entry.Level, entry.Message)
	
	if entry.Component != "" {
		result += fmt.Sprintf(" component=%s", entry.Component)
	}

	if entry.Error != "" {
		result += fmt.Sprintf(" error=\"%s\"", entry.Error)
	}

	for key, value := range entry.Fields {
		result += fmt.Sprintf(" %s=%v", key, value)
	}

	return result
}

// levelToString converts log level to string
func (l *Logger) levelToString(level LogLevel) string {
	switch level {
	case DebugLevel:
		return "DEBUG"
	case InfoLevel:
		return "INFO"
	case WarnLevel:
		return "WARN"
	case ErrorLevel:
		return "ERROR"
	default:
		return "UNKNOWN"
	}
}

// parseLogLevel parses log level from string
func parseLogLevel(level string) LogLevel {
	switch level {
	case "debug":
		return DebugLevel
	case "info":
		return InfoLevel
	case "warn", "warning":
		return WarnLevel
	case "error":
		return ErrorLevel
	default:
		return InfoLevel
	}
}

// parseLogFormat parses log format from string
func parseLogFormat(format string) LogFormat {
	switch format {
	case "json":
		return JSONFormat
	case "text":
		return TextFormat
	default:
		return JSONFormat
	}
}

// LogEntry represents a log entry
type LogEntry struct {
	Timestamp string                 `json:"timestamp"`
	Level     string                 `json:"level"`
	Message   string                 `json:"message"`
	Component string                 `json:"component,omitempty"`
	Error     string                 `json:"error,omitempty"`
	Fields    map[string]interface{} `json:"fields,omitempty"`
}

// WithComponent creates a new logger with a specific component name
func (l *Logger) WithComponent(component string) ports.Logger {
	return &Logger{
		level:     l.level,
		format:    l.format,
		output:    l.output,
		component: component,
	}
}

// SetLevel changes the logging level
func (l *Logger) SetLevel(level string) {
	l.level = parseLogLevel(level)
}

// SetFormat changes the logging format
func (l *Logger) SetFormat(format string) {
	l.format = parseLogFormat(format)
}