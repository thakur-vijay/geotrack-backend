package app

import (
	"database/sql"
	"net/http"

	"github.com/thakur-vijay/geotrack-backend/internal/api"
	"github.com/thakur-vijay/geotrack-backend/internal/middleware"
	"github.com/thakur-vijay/geotrack-backend/internal/modules/auth"
	"github.com/thakur-vijay/geotrack-backend/internal/modules/groups"
	"github.com/thakur-vijay/geotrack-backend/internal/modules/sessions"
	"github.com/thakur-vijay/geotrack-backend/internal/modules/users"
	"github.com/thakur-vijay/geotrack-backend/internal/realtime"
)

func BuildRouter(db *sql.DB) http.Handler {
	// Sessions
	sessionRepository := sessions.NewRepository(db)
	sessionService := sessions.NewService(sessionRepository)

	// Repositories
	userRepository := users.NewRepository(db)
	groupsRepository := groups.NewRepository(db)

	// Realtime
	realtimeManager := realtime.NewManager()
	// realtimeHandler := realtime.NewHandler(realtimeManager)
	// realtimePublisher := realtime.NewPublisher(realtimeManager)

	// Auth
	authService := auth.NewService(userRepository, sessionService)
	authHandler := auth.NewHandler(authService, sessionService, realtimeManager)

	//Middlewares
	authMiddleware := middleware.NewAuthMiddleware(sessionService)
	clientIdMiddleware := middleware.NewClientIDMiddleware()

	//services
	// Handlers
	userHandler := users.NewHandler(userRepository)
	groupsHandler := groups.NewHandler(groupsRepository)

	// Router
	router := api.NewRouter(api.Dependencies{
		AuthHandler:        authHandler,
		AuthMiddleware:     authMiddleware,
		ClientIdMiddleware: clientIdMiddleware,
		UserHandler:        userHandler,
		GroupsHandler:      groupsHandler,
	})

	return router
}
