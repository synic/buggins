//go:build debug

package store

func init() {
	runMigrations = false
}
