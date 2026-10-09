package groups

import (
	"net/http"

	"github.com/thakur-vijay/geotrack-backend/internal/middleware"
)

func RegisterRoutes(
	mux *http.ServeMux,
	handler *Handler,
	authMiddleware *middleware.Middleware,
) {
	mux.Handle("POST /groups", authMiddleware.RequireAuth(http.HandlerFunc(handler.Create)))
	mux.Handle("GET /groups/{id}", authMiddleware.RequireAuth(http.HandlerFunc(handler.GetByID)))
	mux.Handle("GET /groups", authMiddleware.RequireAuth(http.HandlerFunc(handler.GetMyGroups)))
	mux.Handle("POST /groups/{id}/members", authMiddleware.RequireAuth(http.HandlerFunc(handler.AddMember)))
	mux.Handle("DELETE /groups/{id}/members/{userID}", authMiddleware.RequireAuth(http.HandlerFunc(handler.RemoveMember)))
	mux.Handle("DELETE /groups/{id}", authMiddleware.RequireAuth(http.HandlerFunc(handler.Delete)))
}
