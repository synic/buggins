//go:build release

package store

func init() {
	runMigrations = true
}
