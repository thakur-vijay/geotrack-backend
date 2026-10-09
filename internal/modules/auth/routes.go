package auth

import (
	"net/http"

	"github.com/thakur-vijay/geotrack-backend/internal/middleware"
)

func RegisterRoutes(
	mux *http.ServeMux,
	handler *Handler,
	authMiddleware *middleware.Middleware,
	clientIDMiddleware *middleware.ClientIDMiddleware,
) {
	mux.HandleFunc("POST /auth/signup", handler.Signup)
	mux.HandleFunc("POST /auth/login", handler.Login)
	mux.HandleFunc("POST /auth/refresh", handler.Refresh)
	mux.Handle(
		"POST /auth/logout",
		authMiddleware.RequireAuth(clientIDMiddleware.RequireClientID(http.HandlerFunc(handler.Logout))),
	)
}
