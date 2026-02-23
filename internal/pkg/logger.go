package pkg

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

type Logger struct {
	logPath string
}

func NewLogger(projectRoot string) (*Logger, error) {
	logDir := filepath.Join(projectRoot, ".aerostack", "logs")
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create log directory: %w", err)
	}

	return &Logger{
		logPath: filepath.Join(logDir, "cli.log"),
	}, nil
}

func (l *Logger) Log(level, message string) {
	f, err := os.OpenFile(l.logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return
	}
	defer f.Close()

	timestamp := time.Now().Format(time.RFC3339)
	entry := fmt.Sprintf("[%s] [%s] %s\n", timestamp, level, message)
	f.WriteString(entry)
}

func (l *Logger) Error(message string, err error) {
	msg := message
	if err != nil {
		msg = fmt.Sprintf("%s: %v", message, err)
	}
	l.Log("ERROR", msg)
}

func (l *Logger) Info(message string) {
	l.Log("INFO", message)
}

func (l *Logger) GetLogContent() (string, error) {
	content, err := os.ReadFile(l.logPath)
	if err != nil {
		return "", err
	}
	return string(content), nil
}
