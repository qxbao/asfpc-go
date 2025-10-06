package logger

import (
	"testing"
)

func TestInitLogger(t *testing.T) {
	err := InitLogger(true)
	if err != nil {
		t.Fatalf("Failed to initialize logger in development mode: %v", err)
	}
	err = InitLogger(false)
	if err != nil {
		t.Fatalf("Failed to initialize logger in production mode: %v", err)
	}
	if Logger == nil {
		t.Fatal("Logger is nil after initialization")
	}
	logger := GetLogger("test")
	if logger == nil {
		t.Fatal("GetLogger returned nil")
	}
	err = FlushLogger()
	if err != nil {
		t.Fatalf("Failed to flush logger: %v", err)
	}
	t.Run("Flush nil logger", func(t *testing.T) {
		Logger = nil
		defer func() {
			if r := recover(); r == nil {
				t.Error("Expected panic when flushing nil logger, but did not panic")
			}
		}()
		_ = FlushLogger()
		if err == nil {
			t.Error("Expected error when flushing nil logger, got nil")
		}
	})
}