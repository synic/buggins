//go:build mage

package main

import (
	"errors"
	"fmt"
	"os"
	"path"
	"path/filepath"

	"github.com/magefile/mage/mg"
	"github.com/magefile/mage/sh"
)

var (
	Default = Dev

	// paths
	binPath     = "bin"
	debugPath   = path.Join(binPath, "bot-debug")
	releasePath = path.Join(binPath, "bot-release")
	runCmd      = sh.RunCmd("go", "run")
	buildCmd    = sh.RunCmd("go", "build")

	// required command line tools (versions are specified in go.mod)
	tools = map[string]string{
		"air":         "github.com/air-verse/air",
		"goose":       "github.com/pressly/goose/v3/cmd/goose",
		"sqlc":        "github.com/sqlc-dev/sqlc/cmd/sqlc",
		"staticcheck": "honnef.co/go/tools/cmd/staticcheck",
	}

	// aliases
	P = filepath.FromSlash
)

func Dev() error {
	mg.Deps(Deps.Dev)

	return sh.RunV(path.Join(binPath, "air"))
}

type Deps mg.Namespace

func (Deps) Dev() error {
	gobin, err := filepath.Abs(binPath)

	if err != nil {
		return err
	}

	for name, location := range tools {
		_, err = os.Stat(path.Join(binPath, name))

		if err == nil {
			continue
		} else if !errors.Is(err, os.ErrNotExist) {
			return err
		}

		fmt.Printf("installing tool %s ...\n", location)
		err = sh.RunWithV(map[string]string{"GOBIN": gobin}, "go", "install", location)

		if err != nil {
			return err
		}
	}

	return nil
}

type Build mg.Namespace

func (Build) Dev() error {
	mg.Deps(Codegen)

	return buildCmd("-race", "-tags", "debug", "-o", debugPath, ".")
}

func (Build) Release() error {
	return buildCmd("-tags", "release", "-ldflags", "\"-s -w\"", "-o", releasePath, ".")
}

func Codegen() error {
	mg.Deps(Deps.Dev)

	return sh.RunV("go", "generate", "-n", "./...")
}

func Lint() error {
	mg.Deps(Deps.Dev)

	err := sh.Run("go", "vet", "./...")

	if err != nil {
		return err
	}

	return sh.RunV(path.Join(binPath, "staticcheck"), "./...")
}

func Test() error {
	return sh.Run("go", "test", "-race", "./...")
}

func Clean() error {
	files, err := os.ReadDir(binPath)

	if err != nil {
		return err
	}

	for _, file := range files {
		err := sh.Rm(path.Join(binPath, file.Name()))

		if err != nil {
			return err
		}
	}

	return nil
}

func migrationEnv() map[string]string {
	return map[string]string{
		"GOOSE_MIGRATION_DIR": P("internal/store/migrations"),
		"GOOSE_DRIVER":        "sqlite3",
		"GOOSE_DBSTRING":      "db.sqlite",
	}
}

type Migrate mg.Namespace

func (Migrate) Up() error {
	return sh.RunWithV(migrationEnv(), "goose", "up")
}

func (Migrate) Down() error {
	return sh.RunWithV(migrationEnv(), "goose", "down")
}

func (Migrate) Create(name string) error {
	return sh.RunWithV(migrationEnv(), "goose", "create", name, "sql")
}
