package api

import (
	"net/http"

	"github.com/thakur-vijay/geotrack-backend/internal/middleware"
	"github.com/thakur-vijay/geotrack-backend/internal/modules/auth"
	"github.com/thakur-vijay/geotrack-backend/internal/modules/groups"
	"github.com/thakur-vijay/geotrack-backend/internal/modules/regions"
	"github.com/thakur-vijay/geotrack-backend/internal/modules/users"
	"github.com/thakur-vijay/geotrack-backend/internal/realtime"
)

const v1Prefix = "/api/v1"

type Dependencies struct {
	AuthHandler    *auth.Handler
	AuthMiddleware *middleware.Middleware
	UserHandler    *users.Handler
	GroupsHandler  *groups.Handler
	RegionsHandler *regions.Handler
	RealtimeHander *realtime.Handler
}

func NewRouter(dependencies Dependencies) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", health)

	v1Mux := http.NewServeMux()
	auth.RegisterRoutes(v1Mux, dependencies.AuthHandler, dependencies.AuthMiddleware)
	users.RegisterRoutes(v1Mux, dependencies.UserHandler, dependencies.AuthMiddleware)
	groups.RegisterRoutes(v1Mux, dependencies.GroupsHandler, dependencies.AuthMiddleware)
	regions.RegisterRoutes(v1Mux, dependencies.RegionsHandler, dependencies.AuthMiddleware)
	realtime.RegisterRoutes(v1Mux, dependencies.RealtimeHander, dependencies.AuthMiddleware)
	mux.Handle(v1Prefix+"/", http.StripPrefix(v1Prefix, v1Mux))

	return mux
}

func health(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(`{"status":"ok"}`))
}
