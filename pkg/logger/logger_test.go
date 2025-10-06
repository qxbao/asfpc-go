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
		// On Windows and some CI environments, syncing stderr/stdout returns "invalid argument"
		// This is expected behavior and not a real error
		errMsg := err.Error()
		if errMsg != "sync /dev/stderr: invalid argument" &&
			errMsg != "sync /dev/stdout: invalid argument" &&
			errMsg != "sync /dev/stderr: The handle is invalid." &&
			errMsg != "sync /dev/stdout: The handle is invalid." {
			t.Fatalf("Failed to flush logger: %v", err)
		}
		// Log but don't fail - this is expected on Windows/CI
		t.Logf("Flush logger warning (expected on Windows/CI): %v", err)
	}
	Logger = nil
	if FlushLogger() != nil {
		t.Fatal("FlushLogger should return nil when Logger is nil")
	}
}
