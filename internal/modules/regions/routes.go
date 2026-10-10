package regions

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
		"POST /safe-regions",
		authMiddleware.RequireAuth(http.HandlerFunc(handler.Create)),
	)

	mux.Handle(
		"GET /safe-regions",
		authMiddleware.RequireAuth(http.HandlerFunc(handler.Fetch)),
	)

	mux.Handle(
		"GET /safe-regions/{id}",
		authMiddleware.RequireAuth(http.HandlerFunc(handler.FetchByID)),
	)

	mux.Handle(
		"PUT /safe-regions/{id}",
		authMiddleware.RequireAuth(http.HandlerFunc(handler.Update)),
	)

	mux.Handle(
		"DELETE /safe-regions/{id}",
		authMiddleware.RequireAuth(http.HandlerFunc(handler.Delete)),
	)

	mux.Handle(
		"GET /groups/{group_id}/safe-regions",
		authMiddleware.RequireAuth(http.HandlerFunc(handler.FetchByGroupID)),
	)
}
