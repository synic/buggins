//go:build mage

package main

import (
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

	// aliases
	P = filepath.FromSlash

	// commands
	buildCmd       = sh.RunCmd("go", "build")
	airCmd         = sh.RunCmd("go", "tool", "github.com/air-verse/air")
	staticCheckCmd = sh.RunCmd("go", "tool", "honnef.co/go/tools/cmd/staticcheck")
)

func Dev() error {
	return airCmd()
}

type Build mg.Namespace

func (Build) Dev() error {
	return buildCmd("-tags", "debug", "-o", debugPath, ".")
}

func (Build) Release() error {
	env := map[string]string{"CGO_ENABLED": "1"}
	return sh.RunWithV(
		env,
		"go",
		"build",
		"-a",
		"-tags",
		"release",
		"-ldflags",
		"-s -w -linkmode external -extldflags \"-static\"",
		"-o",
		releasePath,
		".",
	)
}

func Codegen() error {
	return sh.Run("go", "generate", "./...")
}

func Lint() error {
	err := sh.Run("go", "vet", "./...")

	if err != nil {
		return err
	}

	return staticCheckCmd()
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
		if file.Name() == ".gitkeep" {
			continue
		}

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
	return sh.RunWithV(migrationEnv(), "go", "tool", "github.com/pressly/goose/v3/cmd/goose", "up")
}

func (Migrate) Down() error {
	return sh.RunWithV(
		migrationEnv(),
		"go",
		"tool",
		"github.com/pressly/goose/v3/cmd/goose",
		"down",
	)
}

func (Migrate) Create(name string) error {
	return sh.RunWithV(
		migrationEnv(),
		"go",
		"tool",
		"github.com/pressly/goose/v3/cmd/goose",
		"create",
		name,
		"sql",
	)
}
