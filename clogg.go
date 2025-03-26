package clogg

import (
	"context"
	"log/slog"
	"os"
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
func New(bufferSize int) *Logger {
	handler := slog.NewJSONHandler(os.Stdout, nil)
	slogger := slog.New(handler)

	clogger := &Logger{
		logChan: make(chan log, bufferSize),
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
	for msg := range l.logChan {
		// Process each log message and output it using the slog.Logger.
		l.logger.LogAttrs(context.Background(), msg.Level, msg.Msg, msg.Attrs...)
	}
	// Signal that the log processing is complete by closing the done channel.
	close(l.done)
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

// Debug logs a debug message
func (l *Logger) Debug(ctx context.Context, msg string, attrs ...slog.Attr) {
	select {
	case l.logChan <- log{Level: slog.LevelDebug, Msg: msg, Attrs: attrs}:
		// enqueued successfully
	default:
		// Optional: drop or handle full buffer
	}
}

// Info logs an info message
func (l *Logger) Info(ctx context.Context, msg string, attrs ...slog.Attr) {
	select {
	case l.logChan <- log{Level: slog.LevelInfo, Msg: msg, Attrs: attrs}:
		// enqueued successfully
	default:
		// Optional: drop or handle full buffer
	}
}

// Warn logs a warning message
func (l *Logger) Warn(ctx context.Context, msg string, attrs ...slog.Attr) {
	select {
	case l.logChan <- log{Level: slog.LevelWarn, Msg: msg, Attrs: attrs}:
		// enqueued successfully
	default:
		// Optional: drop or handle full buffer
	}
}

// Error logs an error message
func (l *Logger) Error(ctx context.Context, msg string, attrs ...slog.Attr) {
	select {
	case l.logChan <- log{Level: slog.LevelError, Msg: msg, Attrs: attrs}:
		// enqueued successfully
	default:
		// Optional: drop or handle full buffer
	}
}
