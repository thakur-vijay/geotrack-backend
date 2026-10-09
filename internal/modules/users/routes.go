package users

import (
	"net/http"

	"github.com/thakur-vijay/geotrack-backend/internal/middleware"
)

func RegisterRoutes(
	mux *http.ServeMux,
	handler *Handler,
	authMiddleware *middleware.Middleware,
) {
	mux.Handle(
		"GET /users/{id}",
		authMiddleware.RequireAuth(http.HandlerFunc(handler.GetByID)),
	)
}
