package app

import (
	"database/sql"
	"net/http"

	"github.com/thakur-vijay/geotrack-backend/internal/api"
	"github.com/thakur-vijay/geotrack-backend/internal/middleware"
	"github.com/thakur-vijay/geotrack-backend/internal/modules/auth"
	"github.com/thakur-vijay/geotrack-backend/internal/modules/groups"
	"github.com/thakur-vijay/geotrack-backend/internal/modules/regions"
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
	regionsRepository := regions.NewRepository(db)

	// Realtime
	realtimeManager := realtime.NewManager()
	// realtimeHandler := realtime.NewHandler(realtimeManager)
	// realtimePublisher := realtime.NewPublisher(realtimeManager)

	// Auth
	authService := auth.NewService(userRepository, sessionService)
	authHandler := auth.NewHandler(authService, sessionService)

	//Middlewares
	authMiddleware := middleware.NewAuthMiddleware(sessionService)
	//services
	// Handlers
	userHandler := users.NewHandler(userRepository)
	groupsHandler := groups.NewHandler(groupsRepository)
	regionsHandler := regions.NewHandler(regionsRepository)
	realtimeHandler := realtime.NewHandler(realtimeManager, groupsRepository)

	// Router
	router := api.NewRouter(api.Dependencies{
		AuthHandler:    authHandler,
		AuthMiddleware: authMiddleware,
		UserHandler:    userHandler,
		GroupsHandler:  groupsHandler,
		RegionsHandler: regionsHandler,
		RealtimeHander: realtimeHandler,
	})

	return router
}
