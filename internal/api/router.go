package api

import (
	"net/http"

	"github.com/thakur-vijay/geotrack-backend/internal/middleware"
	"github.com/thakur-vijay/geotrack-backend/internal/modules/auth"
	"github.com/thakur-vijay/geotrack-backend/internal/modules/users"
)

const v1Prefix = "/api/v1"

type Dependencies struct {
	AuthHandler        *auth.Handler
	AuthMiddleware     *middleware.Middleware
	ClientIdMiddleware *middleware.ClientIDMiddleware
	UserHandler        *users.Handler
}

func NewRouter(dependencies Dependencies) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", health)

	v1Mux := http.NewServeMux()
	auth.RegisterRoutes(v1Mux, dependencies.AuthHandler, dependencies.AuthMiddleware, dependencies.ClientIdMiddleware)
	users.RegisterRoutes(v1Mux, dependencies.UserHandler, dependencies.AuthMiddleware)
	mux.Handle(v1Prefix+"/", http.StripPrefix(v1Prefix, v1Mux))

	return mux
}

func health(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(`{"status":"ok"}`))
}
