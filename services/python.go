package services

import (
	"os/exec"
	"path/filepath"
	"fmt"
)

const PythonPath = "services/python"

type PythonService struct {
	EnvName string
}


func (ps PythonService) RunScript(args ...string) (string, error) {
	pythonExe := filepath.Join("venv", "Scripts", "python.exe")
	cmdArgs := append([]string{"main.py"}, args...)
	
	cmd := exec.Command(pythonExe, cmdArgs...)
	cmd.Dir = PythonPath

	output, err := cmd.CombinedOutput()
	if err != nil {
		return string(output), fmt.Errorf("Python script failed: %v\nOutput: %s", err, string(output))
	}

	return string(output), nil
}
