package users

import (
	"database/sql"
	"log"
	"net/http"

	"github.com/thakur-vijay/geotrack-backend/internal/response"
)

type Handler struct {
	repository *Repository
}

func NewHandler(
	repository *Repository,
) *Handler {
	return &Handler{
		repository: repository,
	}
}

func (h *Handler) GetByID(w http.ResponseWriter, r *http.Request) {
	userID := r.PathValue("id")

	user, err := h.repository.GetByID(r.Context(), userID)
	if err != nil {
		if err == sql.ErrNoRows {
			response.Write(w, http.StatusNotFound, "user was not found")
			return
		}

		log.Printf("GetByID error: %v", err)
		response.Write(w, http.StatusInternalServerError, "user could not be retrieved")
		return
	}
	response.WriteItem(w, http.StatusOK, "user retrieved successfully", user)
}
