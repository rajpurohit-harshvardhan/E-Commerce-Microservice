package migrate

import (
	"database/sql"
	"embed"
	"fmt"
	"os"

	"github.com/pressly/goose/v3"
)

// Run executes migrations using goose and an embedded FS.
// - db: the *sql.DB connection
// - fs: the embed.FS with migration files
// - dir: the virtual directory inside the FS (usually ".")
func Run(db *sql.DB, fs embed.FS, dir string) error {
	goose.SetBaseFS(fs)
	if err := goose.SetDialect("postgres"); err != nil {
		return err
	}

	action := os.Getenv("MIGRATION_ACTION") // defaults to up

	switch action {
	case "", "up":
		return goose.Up(db, dir)
	case "down":
		return goose.Down(db, dir)
	case "redo":
		return goose.Redo(db, dir)
	case "status":
		return goose.Status(db, dir)
	//case "up-to":
	//	v := os.Getenv("MIGRATION_VERSION")
	//	if v == "" {
	//		return errors.New("MIGRATION_VERSION required for up-to")
	//	}
	//	return goose.UpTo(db, dir, mustParseInt64(v))
	//case "down-to":
	//	v := os.Getenv("MIGRATION_VERSION")
	//	if v == "" {
	//		return errors.New("MIGRATION_VERSION required for down-to")
	//	}
	//	return goose.DownTo(db, dir, mustParseInt64(v))
	default:
		return fmt.Errorf("unknown MIGRATION_ACTION: %s", action)
	}
}
