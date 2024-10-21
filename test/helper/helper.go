package helper

import (
	"os/exec"
	"path/filepath"
	"runtime"
)

func BuildPrinterExecutable() (string, error) {
	filename := "printer"
	if runtime.GOOS == "windows" {
		filename = "printer.exe"
	}
	cmd := exec.Command("go", "build", "-o", filename)
	cmd.Dir = filepath.Join("..", "helper", "printer")
	err := cmd.Run()
	if err != nil {
		return "", err
	}
	return filepath.Abs(filepath.Join(cmd.Dir, filename))
}
