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
	Logger = nil
	if FlushLogger() != nil {
		t.Fatal("FlushLogger should return nil when Logger is nil")
	}
}