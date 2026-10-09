package database

import (
	"database/sql"
	"fmt"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func Migrate(db *sql.DB, databaseURL string) error {
	if databaseURL == "" {
		return fmt.Errorf("DATABASE_URL is not set")
	}

	m, err := migrate.New(
		"file://internal/database/migrations",
		databaseURL,
	)
	if err != nil {
		return err
	}
	defer m.Close()

	err = m.Up()
	if err != nil && err != migrate.ErrNoChange {
		return err
	}

	return nil
}
