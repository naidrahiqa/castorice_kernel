package logger

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"
)

var (
	logFile   *os.File
	stdLogger *log.Logger
)

// Init initializes the logger and returns the log file handle (caller should defer Close)
func Init() (*os.File, error) {
	appData := os.Getenv("APPDATA")
	if appData == "" {
		appData = "."
	}

	logDir := filepath.Join(appData, "KernelRescueTool")
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create log directory: %w", err)
	}

	logPath := filepath.Join(logDir, "app.log")
	f, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to open log file: %w", err)
	}

	logFile = f
	stdLogger = log.New(f, "", 0)

	Info("Logger initialized — log file: %s", logPath)
	return f, nil
}

// Info logs an info-level message
func Info(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	entry := fmt.Sprintf("[%s] [INFO]  %s", timestamp(), msg)
	if stdLogger != nil {
		stdLogger.Println(entry)
	}
	fmt.Println(entry)
}

// Warn logs a warning-level message
func Warn(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	entry := fmt.Sprintf("[%s] [WARN]  %s", timestamp(), msg)
	if stdLogger != nil {
		stdLogger.Println(entry)
	}
	fmt.Println(entry)
}

// Error logs an error-level message
func Error(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	entry := fmt.Sprintf("[%s] [ERROR] %s", timestamp(), msg)
	if stdLogger != nil {
		stdLogger.Println(entry)
	}
	fmt.Fprintln(os.Stderr, entry)
}

func timestamp() string {
	return time.Now().Format("2006-01-02 15:04:05")
}
