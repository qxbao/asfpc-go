//go:build integration
// +build integration

package python

import (
	"path"
	"runtime"
	"testing"
)

// This test requires the python virtual environment to be set up with the test task.
func TestRunScript(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	expectedOutput := "Test task executed\r\n"
	_, filename, _, _ := runtime.Caller(0)
	t.Logf("Current test filename: %s", filename)
	pythonPath := path.Join(filename, "..", "..", "..", "..", "python")
	ps := NewPythonService("venv", false, true, &pythonPath)
	output, err := ps.RunScript("--task=test")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if output != expectedOutput {
		t.Fatalf("Expected output to be %q, got %q", expectedOutput, output)
	}
}
