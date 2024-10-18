//go:build mage

package main

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path"
	"path/filepath"

	"github.com/magefile/mage/mg"
	"github.com/magefile/mage/sh"
)

var (
	Default = Dev

	// paths
	binPath  = "bin"
	execPath = path.Join(binPath, "bot")
	runCmd   = sh.RunCmd("go", "run")
	buildCmd = sh.RunCmd("go", "build")

	// required command line tools (versions are specified in go.mod)
	tools = map[string]tool{
		"air":                {path: "github.com/air-verse/air", global: false},
		"goose":              {path: "github.com/pressly/goose/v3/cmd/goose", global: false},
		"sqlc":               {path: "github.com/sqlc-dev/sqlc/cmd/sqlc", global: false},
		"staticcheck":        {path: "honnef.co/go/tools/cmd/staticcheck", global: true},
		"protoc-gen-go":      {path: "google.golang.org/protobuf/cmd/protoc-gen-go", global: true},
		"protoc-gen-go-grpc": {path: "google.golang.org/protobuf/cmd/protoc-gen-go", global: true},
	}

	// aliases
	P = filepath.FromSlash
)

type tool struct {
	path   string
	global bool
}

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

	for name, info := range tools {
		if info.global {
			_, err := exec.LookPath(name)

			if err != nil {
				err = sh.RunV("go", "install", info.path)

				if err != nil {
					return err
				}
			}
			continue
		}

		_, err = os.Stat(path.Join(binPath, name))

		if err == nil {
			continue
		} else if !errors.Is(err, os.ErrNotExist) {
			return err
		}

		fmt.Printf("installing tool %s ...\n", info.path)
		err = sh.RunWithV(map[string]string{"GOBIN": gobin}, "go", "install", info.path)

		if err != nil {
			return err
		}
	}

	return nil
}

type Build mg.Namespace

func (Build) Dev() error {
	return buildCmd("-tags", "debug", "-o", execPath, ".")
}

func (Build) Release() error {
	return buildCmd("-tags", "release", "-ldflags", "\"-s -w\"", "-o", execPath, ".")
}

func Codegen() error {
	mg.Deps(Deps.Dev)

	return sh.Run("go", "generate", "./...")
}

func Lint() error {
	mg.Deps(Deps.Dev)

	err := sh.Run("go", "vet", "./...")

	if err != nil {
		return err
	}

	return sh.RunV("staticcheck", "./...")
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
	return sh.RunWithV(migrationEnv(), "goose", "up")
}

func (Migrate) Down() error {
	return sh.RunWithV(migrationEnv(), "goose", "down")
}

func (Migrate) Create(name string) error {
	return sh.RunWithV(migrationEnv(), "goose", "create", name, "sql")
}
