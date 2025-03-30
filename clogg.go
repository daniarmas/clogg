package clogg

import (
	"context"
	"fmt"
	"log/slog"
	"time"
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

type LoggerConfig struct {
    BufferSize int
    Handler    slog.Handler
}

// New creates a new asynchronous Logger instance with the specified buffer size.
//
// Parameters:
// - bufferSize: The size of the buffered channel used to enqueue log messages for asynchronous processing.
//
// Behavior:
//   - The function initializes a new Logger with a buffered channel (`logChan`) to hold log messages,
//     an underlying `slog.Logger` for formatting and outputting logs, and a `done` channel to signal
//     when the log processing goroutine has finished.
//   - It starts a background goroutine (`processLogs`) to process log messages asynchronously.
func New(config LoggerConfig) *Logger {
	slogger := slog.New(config.Handler)

	clogger := &Logger{
		logChan: make(chan log, config.BufferSize),
		logger:  slogger,
		done:    make(chan struct{}),
	}

	go clogger.processLogs()
	return clogger
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
	close(l.logChan)
	<-l.done // wait until log processor finishes
}

// log is a method that enqueues a log message with the specified level, message, and attributes.
func (l *Logger) log(ctx context.Context, level slog.Level, msg string, attrs ...slog.Attr) error {
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
	err := l.log(ctx, level, msg, attrs...)
	if err != nil {
		// Log an error if the message could not be enqueued
		l.logger.LogAttrs(ctx, slog.LevelError, "failed to log", slog.String("error", err.Error()))
	}
}

// Debug logs a debug message
func (l *Logger) Debug(ctx context.Context, msg string, attrs ...slog.Attr) {
	l.logWithLevel(ctx, slog.LevelDebug, msg, attrs...)
}

// Info logs an info message
func (l *Logger) Info(ctx context.Context, msg string, attrs ...slog.Attr) {
	l.logWithLevel(ctx, slog.LevelInfo, msg, attrs...)
}

// Warn logs a warning message
func (l *Logger) Warn(ctx context.Context, msg string, attrs ...slog.Attr) {
	l.logWithLevel(ctx, slog.LevelWarn, msg, attrs...)
}

// Error logs an error message
func (l *Logger) Error(ctx context.Context, msg string, attrs ...slog.Attr) {
	l.logWithLevel(ctx, slog.LevelError, msg, attrs...)
}
