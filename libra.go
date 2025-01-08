package main

import (
	"os"
	"path/filepath"

	"github.com/libramusic/libracore/cmds"
)

func main() {
	executablePath, _ := os.Executable()
	executablePath, _ = filepath.EvalSymlinks(executablePath)
	_ = os.Chdir(filepath.Dir(executablePath))

	_ = cmds.Execute()
}
