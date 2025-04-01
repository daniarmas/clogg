package clogg

import (
	"context"
	"log/slog"
	"os"
	"sync"
	"testing"
	"time"
)

func TestGetLogger_DefaultConfig(t *testing.T) {
	logger := GetLogger(LoggerConfig{})
	if logger == nil {
		t.Fatal("Expected logger to be initialized, got nil")
	}
	if globalLogger != logger {
		t.Fatal("Expected globalLogger to match the returned logger")
	}
}

func TestGetLogger_CustomConfig(t *testing.T) {
	handler := slog.NewJSONHandler(os.Stdout, nil)
	logger := GetLogger(LoggerConfig{BufferSize: 50, Handler: handler})
	if logger == nil {
		t.Fatal("Expected logger to be initialized, got nil")
	}
	if cap(logger.logChan) != 50 {
		t.Fatalf("Expected logChan buffer size to be 50, got %d", cap(logger.logChan))
	}
}

func TestLogger_Singleton(t *testing.T) {
	wg := sync.WaitGroup{}
	wg.Add(2)

	var logger1, logger2 *Logger
	go func() {
		defer wg.Done()
		logger1 = GetLogger(LoggerConfig{})
	}()
	go func() {
		defer wg.Done()
		logger2 = GetLogger(LoggerConfig{})
	}()
	wg.Wait()

	if logger1 != logger2 {
		t.Fatal("Expected logger to be a singleton")
	}
}

func TestLogger_Debug(t *testing.T) {
	logger := GetLogger(LoggerConfig{})
	ctx := context.Background()
	logger.debug(ctx, "debug message", String("key", "value"))

	select {
	case msg := <-logger.logChan:
		if msg.Level != slog.LevelDebug {
			t.Fatalf("Expected level Debug, got %v", msg.Level)
		}
		if msg.Msg != "debug message" {
			t.Fatalf("Expected message 'debug message', got %s", msg.Msg)
		}
	default:
		t.Fatal("Expected message in logChan, got none")
	}
}

func TestLoggerShutdown(t *testing.T) {
	// Create a logger with a small buffer size for testing
	logger := GetLogger(LoggerConfig{
		BufferSize: 2,
		Handler:    nil, // Replace with a mock handler if needed
	})

	// Log some messages
	Debug(context.Background(), "Test message 1")
	Debug(context.Background(), "Test message 2")

	// Call Shutdown
	logger.Shutdown()

	// Verify that the logger's done channel is closed
	select {
	case <-logger.done:
		// Success: done channel is closed
	default:
		t.Fatal("Logger did not shut down properly: done channel is not closed")
	}
}

func TestLogger_BufferFull(t *testing.T) {
	handler := slog.NewJSONHandler(os.Stdout, nil)
	logger := GetLogger(LoggerConfig{BufferSize: 1, Handler: handler})
	ctx := context.Background()

	logger.debug(ctx, "first message")
	logger.debug(ctx, "second message") // This should trigger retry logic

	time.Sleep(50 * time.Millisecond) // Allow retries to occur

	select {
	case msg := <-logger.logChan:
		if msg.Msg != "first message" {
			t.Fatalf("Expected 'first message', got %s", msg.Msg)
		}
	default:
		t.Fatal("Expected message in logChan, got none")
	}
}

func TestAttrHelpers(t *testing.T) {
	attr := String("key", "value")
	if attr.Key != "key" || attr.Value.String() != "value" {
		t.Fatalf("Expected key='key' and value='value', got key='%s' and value='%v'", attr.Key, attr.Value)
	}

	attr = Int("key", 42)
	if attr.Key != "key" || attr.Value.Int64() != 42 {
		t.Fatalf("Expected key='key' and value=42, got key='%s' and value='%v'", attr.Key, attr.Value)
	}

	attr = Bool("key", true)
	if attr.Key != "key" || attr.Value.Bool() != true {
		t.Fatalf("Expected key='key' and value=true, got key='%s' and value='%v'", attr.Key, attr.Value)
	}

	attr = Float64("key", 3.14)
	if attr.Key != "key" || attr.Value.Float64() != 3.14 {
		t.Fatalf("Expected key='key' and value=3.14, got key='%s' and value='%v'", attr.Key, attr.Value)
	}

	now := time.Now().UTC()
	attr = Time("key", now)
	if attr.Key != "key" || attr.Value.Time() != now {
		t.Fatalf("Expected key='key' and value=%v, got key='%s' and value='%v'", now, attr.Key, attr.Value)
	}
}
