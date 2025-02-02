//go:build mage
// +build mage

package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/magefile/mage/mg"
	"github.com/magefile/mage/sh"
)

const binaryName = "libra"

var Default = Build

var Aliases = map[string]interface{}{
	"b":    Build,
	"test": Test.Unit,
	"t":    Test.Unit,
	"tu":   Test.Unit,
	"ti":   Test.Integration,
	"ta":   Test.All,
	"dep":  Deps,
	"d":    Deps,
	"l":    Lint,
	"c":    Clean,
}

func Build() error {
	mg.Deps(Deps)
	fmt.Println("Building...")
	return sh.Run("go", "build", "-v", "-o", binaryName)
}

type Test mg.Namespace

func (Test) Unit() error {
	mg.Deps(Deps)
	fmt.Println("Running unit tests...")
	return sh.Run("go", "test", "./...")
}

func (Test) Integration() error {
	mg.Deps(Deps)
	fmt.Println("Running integration tests...")
	return sh.Run("go", "test", "-tags=integration", "./...")
}

func (Test) All() {
	mg.Deps(Test.Unit, Test.Integration)
}

func Deps() error {
	fmt.Println("Installing dependencies...")
	return sh.Run("go", "mod", "download")
}

func Lint() error {
	fmt.Println("Linting...")

	if _, err := sh.Output("golangci-lint", "version"); err != nil {
		fmt.Println("golangci-lint is not installed. Please install it from https://golangci-lint.run/welcome/install/")
	} else {
		if err = sh.Run("golangci-lint", "run"); err != nil {
			return err
		}
	}

	if _, err := sh.Output("ruff", "version"); err != nil {
		fmt.Println("ruff is not installed. Please install it from https://docs.astral.sh/ruff/installation/")
	} else {
		if err = sh.Run("ruff", "check", "--fix"); err != nil {
			return err
		}
		if err = sh.Run("ruff", "format"); err != nil {
			return err
		}
	}

	if _, err := sh.Output("shellcheck", "--version"); err != nil {
		fmt.Println("shellcheck is not installed. Please install it from https://github.com/koalaman/shellcheck#installing")
	} else {
		if err = sh.RunV("find", ".", "-type", "f", "-name", "*.sh", "-exec", "shellcheck", "{}", "+"); err != nil {
			return err
		}
	}

	return nil
}

func Clean() error {
	fmt.Println("Cleaning...")

	if err := os.Remove(binaryName); err != nil {
		return err
	}

	files, err := filepath.Glob(binaryName + "-*")
	if err != nil {
		return err
	}
	for _, f := range files {
		_ = os.Remove(f)
	}

	if err = os.RemoveAll("dist"); err != nil {
		return err
	}
	if err = os.RemoveAll("completions"); err != nil {
		return err
	}
	if err = os.RemoveAll("manpages"); err != nil {
		return err
	}

	if err = os.RemoveAll(".ruff_cache"); err != nil {
		return err
	}

	return sh.Run("go", "clean")
}
