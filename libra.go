package main

import (
	"os"
	"path/filepath"

	"github.com/LibraMusic/LibraCore/cmds"
)

func main() {
	executablePath, _ := os.Executable()
	executablePath, _ = filepath.EvalSymlinks(executablePath)
	os.Chdir(filepath.Dir(executablePath))

	cmds.Execute()
}
