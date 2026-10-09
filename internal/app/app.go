package app

import (
	"context"
	"database/sql"
	"net/http"

	"github.com/thakur-vijay/geotrack-backend/internal/config"
	"github.com/thakur-vijay/geotrack-backend/internal/database"
)

type App struct {
	DB     *sql.DB
	Router http.Handler
}

func New(cfg *config.Config) (*App, error) {
	if cfg == nil {
		return nil, nil
	}

	db, err := database.Connect(context.Background(), cfg.DatabaseURL)
	if err != nil {
		return nil, err
	}

	if err := database.Migrate(db, cfg.DatabaseURL); err != nil {
		db.Close()
		return nil, err
	}

	router := BuildRouter(db)

	return &App{
		DB:     db,
		Router: router,
	}, nil
}

func (a *App) Close() error {
	if a == nil {
		return nil
	}

	if a.DB != nil {
		return a.DB.Close()
	}

	return nil
}
