package regions

import (
	"database/sql"
	"errors"
	"log"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/thakur-vijay/geotrack-backend/internal/middleware"
	"github.com/thakur-vijay/geotrack-backend/internal/response"
)

type Handler struct {
	repository *Repository
}

func NewHandler(repository *Repository) *Handler {
	return &Handler{
		repository: repository,
	}
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	userID, _ := middleware.UserID(r.Context())

	var request CreateRegionDTO

	if err := response.DecodeJSONAndValidate(r, &request); err != nil {
		response.Write(w, http.StatusBadRequest, err.Error())
		return
	}

	now := time.Now().UTC()

	region := SafeRegion{
		ID:           uuid.New().String(),
		UserID:       userID,
		GroupID:      request.GroupID,
		Name:         request.Name,
		Latitude:     request.Latitude,
		Longitude:    request.Longitude,
		RadiusMeters: request.RadiusMeters,
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	if err := h.repository.Create(r.Context(), region); err != nil {
		log.Printf("Create SafeRegion error: %v", err)
		response.Write(
			w,
			http.StatusInternalServerError,
			"Safe region could not be created",
		)
		return
	}

	response.WriteItem(
		w,
		http.StatusCreated,
		"Safe region created successfully",
		region,
	)
}

func (h *Handler) Fetch(w http.ResponseWriter, r *http.Request) {
	userID, _ := middleware.UserID(r.Context())

	regions, err := h.repository.Fetch(r.Context(), userID)
	if err != nil {
		log.Printf("Fetch SafeRegions error: %v", err)
		response.Write(
			w,
			http.StatusInternalServerError,
			"Safe regions could not be fetched",
		)
		return
	}

	response.WriteList(
		w,
		http.StatusOK,
		"Safe regions fetched successfully",
		regions,
	)
}

func (h *Handler) FetchByID(w http.ResponseWriter, r *http.Request) {
	regionID := r.PathValue("id")
	userID, _ := middleware.UserID(r.Context())

	region, err := h.repository.FetchByID(
		r.Context(),
		regionID,
		userID,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			response.Write(
				w,
				http.StatusNotFound,
				"Safe region was not found",
			)
			return
		}

		log.Printf("Fetch SafeRegion error: %v", err)
		response.Write(
			w,
			http.StatusInternalServerError,
			"Safe region could not be fetched",
		)
		return
	}

	response.WriteItem(
		w,
		http.StatusOK,
		"Safe region fetched successfully",
		region,
	)
}

func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	regionID := r.PathValue("id")
	userID, _ := middleware.UserID(r.Context())

	var request UpdateRegionDTO

	if err := response.DecodeJSONAndValidate(r, &request); err != nil {
		response.Write(w, http.StatusBadRequest, err.Error())
		return
	}

	updatedRegion, err := h.repository.Update(
		r.Context(),
		regionID,
		userID,
		request,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			response.Write(
				w,
				http.StatusNotFound,
				"Safe region was not found",
			)
			return
		}

		log.Printf("Update SafeRegion error: %v", err)
		response.Write(
			w,
			http.StatusInternalServerError,
			"Safe region could not be updated",
		)
		return
	}

	response.WriteItem(
		w,
		http.StatusOK,
		"Safe region updated successfully",
		updatedRegion,
	)
}

func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	regionID := r.PathValue("id")
	userID, _ := middleware.UserID(r.Context())

	if err := h.repository.Delete(r.Context(), regionID, userID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			response.Write(
				w,
				http.StatusNotFound,
				"Safe region was not found",
			)
			return
		}

		log.Printf("Delete SafeRegion error: %v", err)
		response.Write(
			w,
			http.StatusInternalServerError,
			"Safe region could not be deleted",
		)
		return
	}

	response.Write(
		w,
		http.StatusOK,
		"Safe region deleted successfully",
	)
}

func (h *Handler) FetchByGroupID(w http.ResponseWriter, r *http.Request) {
	groupID := r.PathValue("group_id")
	userID, _ := middleware.UserID(r.Context())

	regions, err := h.repository.FetchByGroupID(
		r.Context(),
		groupID,
		userID,
	)
	if err != nil {
		log.Printf("Fetch SafeRegions by GroupID error: %v", err)
		response.Write(
			w,
			http.StatusInternalServerError,
			"Safe regions could not be fetched",
		)
		return
	}

	response.WriteList(
		w,
		http.StatusOK,
		"Safe regions fetched successfully",
		regions,
	)
}
