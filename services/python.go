package services

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
)

var exc, _ = os.Executable()
var PythonPath = filepath.Join(filepath.Dir(exc), "python")

type PythonService struct {
	EnvName string
}

func (ps PythonService) RunScript(args ...string) (string, error) {
	var pythonExe string

	if runtime.GOOS == "windows" {
		pythonExe = filepath.Join("venv", "Scripts", "python.exe")
	} else {
		pythonExe = filepath.Join("venv", "bin", "python")
	}

	args = append(args, "--silent")

	cmdArgs := append([]string{"main.py"}, args...)

	cmd := exec.Command(pythonExe, cmdArgs...)
	cmd.Dir = PythonPath

	output, err := cmd.CombinedOutput()
	if err != nil {
		return string(output), fmt.Errorf("python script failed: %v\nOutput: %s", err, string(output))
	}

	return string(output), nil
}
