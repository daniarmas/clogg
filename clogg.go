package clogg

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"sync"
	"time"
)

var (
	globalLogger *Logger
	once         = &sync.Once{} // Ensure the logger is initialized only once
)

// log represents a single log message with its associated metadata.
// It is used internally by the Logger to enqueue and process log messages.
type log struct {
	// Level specifies the severity level of the log message (e.g., Debug, Info, Warn, Error).
	Level slog.Level

	// Msg is the main log message to be recorded.
	Msg string

	// Attrs contains additional structured attributes or metadata associated with the log message.
	Attrs []slog.Attr
}

// Logger is an asynchronous logger that processes log messages in a separate goroutine.
// It provides methods to log messages at different severity levels (Debug, Info, Warn, Error).
type Logger struct {
	// logChan is a buffered channel used to enqueue log messages for asynchronous processing.
	logChan chan log

	// logger is the underlying slog.Logger used to format and output log messages.
	logger *slog.Logger

	// done is a channel used to signal when the log processing goroutine has finished.
	done chan struct{}
}

// LoggerConfig defines the configuration for the Logger.
//
// Fields:
// - BufferSize: The size of the buffered channel used to enqueue log messages for asynchronous processing.
// - Handler: The slog.Handler used to format and output log messages.
type LoggerConfig struct {
	BufferSize int
	Handler    slog.Handler
}

// GetLogger returns the singleton instance of the Logger.
// If the logger has not been initialized, it initializes it with the provided configuration.
func GetLogger(config LoggerConfig) *Logger {
	once.Do(func() {
		if config.BufferSize == 0 || config.Handler == nil {
			// Provide default configuration if none is provided
			config = LoggerConfig{
				BufferSize: 100,
				Handler:    slog.NewJSONHandler(os.Stdout, nil),
			}
		}
		slogger := slog.New(config.Handler)

		clogger := &Logger{
			logChan: make(chan log, config.BufferSize),
			logger:  slogger,
			done:    make(chan struct{}),
		}

		go clogger.processLogs()
		globalLogger = clogger
	})
	return globalLogger
}

// Shutdown gracefully shuts down the logger by ensuring that all pending log messages
// in the logChan channel are processed before the logger stops.
//
// Steps:
//  1. It closes the logChan channel to signal the processLogs goroutine that no more
//     log messages will be sent.
//  2. It waits for the done channel to be closed, which indicates that the processLogs
//     goroutine has finished processing all remaining log messages and has exited.
//
// This ensures that no log messages are lost during shutdown and that the logger
// shuts down cleanly.
func (l *Logger) Shutdown() {
	if globalLogger != nil {
		close(l.logChan)
		<-l.done // wait until log processor finishes
	}
}

func Debug(ctx context.Context, msg string, attrs ...Attr) {
	if globalLogger != nil {
		globalLogger.debug(ctx, msg, attrs...)
	}
}

func Info(ctx context.Context, msg string, attrs ...Attr) {
	if globalLogger != nil {
		globalLogger.info(ctx, msg, attrs...)
	}
}

func Warn(ctx context.Context, msg string, attrs ...Attr) {
	if globalLogger != nil {
		globalLogger.warn(ctx, msg, attrs...)
	}
}

func Error(ctx context.Context, msg string, attrs ...Attr) {
	if globalLogger != nil {
		globalLogger.error(ctx, msg, attrs...)
	}
}

// processLogs is a goroutine that continuously processes log messages from the logChan channel.
// It reads messages from the channel, formats them using the underlying slog.Logger, and outputs them.
// When the logChan channel is closed, it processes any remaining messages and then closes the done channel
// to signal that the logging process has completed.
func (l *Logger) processLogs() {
	// Signal that the log processing is complete by closing the done channel.
	defer close(l.done)
	for msg := range l.logChan {
		// Process each log message and output it using the slog.Logger.
		l.logger.LogAttrs(context.Background(), msg.Level, msg.Msg, msg.Attrs...)
	}
}

// log is a method that enqueues a log message with the specified level, message, and attributes.
func (l *Logger) logMsg(ctx context.Context, level slog.Level, msg string, attrs ...slog.Attr) error {
	const maxRetries = 3
	for i := 0; i < maxRetries; i++ {
		select {
		case l.logChan <- log{Level: level, Msg: msg, Attrs: attrs}:
			// Enqueued successfully
			return nil
		default:
			// Log a warning on retry
			l.logger.LogAttrs(ctx, slog.LevelWarn, "retrying log message due to full buffer", slog.Int("attempt", i+1))
			time.Sleep(10 * time.Millisecond) // Fixed delay between retries
		}
	}

	// If retries are exhausted, return an error
	return fmt.Errorf("logging buffer channel full")
}

// logWithLevel is a helper method to log messages at a specific level.
func (l *Logger) logWithLevel(ctx context.Context, level slog.Level, msg string, attrs ...slog.Attr) {
	err := l.logMsg(ctx, level, msg, attrs...)
	if err != nil {
		// Log an error if the message could not be enqueued
		l.logger.LogAttrs(ctx, slog.LevelError, "failed to log", slog.String("error", err.Error()))
	}
}

// Debug logs a debug message
func (l *Logger) debug(ctx context.Context, msg string, attrs ...slog.Attr) {
	l.logWithLevel(ctx, slog.LevelDebug, msg, attrs...)
}

// Info logs an info message
func (l *Logger) info(ctx context.Context, msg string, attrs ...slog.Attr) {
	l.logWithLevel(ctx, slog.LevelInfo, msg, attrs...)
}

// Warn logs a warning message
func (l *Logger) warn(ctx context.Context, msg string, attrs ...slog.Attr) {
	l.logWithLevel(ctx, slog.LevelWarn, msg, attrs...)
}

// Error logs an error message
func (l *Logger) error(ctx context.Context, msg string, attrs ...slog.Attr) {
	l.logWithLevel(ctx, slog.LevelError, msg, attrs...)
}

// Attr is an alias for slog.Attr, used to define structured attributes for log messages.
type Attr = slog.Attr

// String creates a new Attr with a string value.
func String(key, value string) Attr {
	return slog.String(key, value)
}

// Int creates a new Attr with an integer value.
func Int(key string, value int) Attr {
	return slog.Int(key, value)
}

// Bool creates a new Attr with a boolean value.
func Bool(key string, value bool) Attr {
	return slog.Bool(key, value)
}

// Float64 creates a new Attr with a float64 value.
func Float64(key string, value float64) Attr {
	return slog.Float64(key, value)
}

// Time creates a new Attr with a time.Time value.
func Time(key string, value time.Time) Attr {
	return slog.Time(key, value)
}
