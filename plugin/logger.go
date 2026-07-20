package plugin

import (
	"context"
	"golang.org/x/exp/jsonrpc2"
	"strings"
)

// LogLevel is the severity of a log message emitted by a plugin.
type LogLevel string

const (
	// LogLevelTrace emits detailed diagnostic information.
	LogLevelTrace LogLevel = "trace"
	// LogLevelDebug emits debugging information.
	LogLevelDebug LogLevel = "debug"
	// LogLevelInfo emits general operational information.
	LogLevelInfo LogLevel = "info"
	// LogLevelWarn emits an unusual condition.
	LogLevelWarn LogLevel = "warn"
	// LogLevelError emits a broken or failed condition.
	LogLevelError LogLevel = "error"
)

// LogNotification is the payload of a plugin log notification.
type LogNotification struct {
	Level   LogLevel `json:"level"`
	Message string   `json:"message"`
}

type Logger struct {
	notifyChan chan<- []byte
}

func NewLogger(notifyChan chan<- []byte) *Logger {
	return &Logger{notifyChan: notifyChan}
}

// Notify emits an arbitrary notification to Core Lightning
func (l *Logger) Notify(ctx context.Context, method string, params any) error {
	notification, err := jsonrpc2.NewNotification(method, params)
	if err != nil {
		return err
	}

	message, err := jsonrpc2.EncodeMessage(notification)
	if err != nil {
		return err
	}

	select {
	case l.notifyChan <- message:
	case <-ctx.Done():
	}

	return nil
}

func (l *Logger) Error(ctx context.Context, message string) {
	l.Log(ctx, message, LogLevelError)
}

func (l *Logger) Warn(ctx context.Context, message string) {
	l.Log(ctx, message, LogLevelWarn)
}

func (l *Logger) Info(ctx context.Context, message string) {
	l.Log(ctx, message, LogLevelInfo)
}

func (l *Logger) Debug(ctx context.Context, message string) {
	l.Log(ctx, message, LogLevelDebug)
}

func (l *Logger) Trace(ctx context.Context, message string) {
	l.Log(ctx, message, LogLevelTrace)
}

// Log emits a log notification to Core Lightning. Each line is emitted as a
// separate notification.
func (l *Logger) Log(ctx context.Context, message string, level LogLevel) {
	for _, line := range strings.Split(message, "\n") {
		err := l.Notify(ctx, "log", LogNotification{
			Level:   level,
			Message: line,
		})

		if err != nil {
			// Should never happen
			panic(err)
		}
	}
}
