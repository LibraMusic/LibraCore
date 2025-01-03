package main

import (
	"os"
	"path/filepath"

	"github.com/charmbracelet/log"

	"github.com/libramusic/libracore/cmds"
)

func main() {
	executablePath, _ := os.Executable()
	executablePath, _ = filepath.EvalSymlinks(executablePath)
	_ = os.Chdir(filepath.Dir(executablePath))

	err := cmds.Execute()
	if err != nil {
		log.Fatal(err)
	}
}
