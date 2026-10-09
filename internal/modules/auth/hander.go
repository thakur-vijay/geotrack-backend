package auth

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"

	"github.com/thakur-vijay/geotrack-backend/internal/middleware"
	"github.com/thakur-vijay/geotrack-backend/internal/modules/sessions"
	"github.com/thakur-vijay/geotrack-backend/internal/response"
)

type Handler struct {
	service            *Service
	sessionService     *sessions.Service
	clientDisconnecter ClientDisconnecter
}

type ClientDisconnecter interface {
	DisconnectClient(userID, clientID string) error
}

func NewHandler(
	service *Service,
	sessionService *sessions.Service,
	clientDisconnecter ClientDisconnecter,
) *Handler {
	return &Handler{
		service:            service,
		sessionService:     sessionService,
		clientDisconnecter: clientDisconnecter,
	}
}

func (h *Handler) Signup(w http.ResponseWriter, r *http.Request) {
	var request SignupRequest

	if err := response.DecodeJSON(r, &request); err != nil {
		response.Write(w, http.StatusBadRequest, err.Error())
		return
	}

	user, accessToken, refreshToken, err := h.service.Signup(
		r.Context(),
		request,
	)
	if err != nil {
		if errors.Is(err, ErrInvalidInput) {
			response.Write(w, http.StatusBadRequest, err.Error())
			return
		}

		log.Printf("Signup error: %v", err)
		response.Write(w, http.StatusInternalServerError, "signup could not be completed")
		return
	}

	payload := LoginResponseDTO{
		User:         user,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}

	response.WriteItem(w, http.StatusCreated, "signup successful", payload)
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var request LoginRequest

	if err := response.DecodeJSON(r, &request); err != nil {
		response.Write(w, http.StatusBadRequest, err.Error())
		return
	}

	user, accessToken, refreshToken, err := h.service.Login(
		r.Context(),
		request,
	)
	if err != nil {
		if errors.Is(err, ErrInvalidCredentials) {
			response.Write(w, http.StatusUnauthorized, "invalid email or password")
			return
		}

		log.Printf("Login error: %v", err)
		response.Write(w, http.StatusInternalServerError, "login could not be completed")
		return
	}

	payload := LoginResponseDTO{
		User:         user,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}

	response.WriteItem(w, http.StatusOK, "login successful", payload)
}

func (h *Handler) Refresh(w http.ResponseWriter, r *http.Request) {
	var request RefreshRequest

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		response.Write(w, http.StatusBadRequest, "request body is invalid")
		return
	}

	accessToken, refreshToken, err := h.sessionService.RefreshSession(
		r.Context(),
		request.RefreshToken,
	)
	if err != nil {
		response.Write(w, http.StatusUnauthorized, "refresh token is invalid")
		return
	}

	payload := RefreshResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}

	response.WriteItem(w, http.StatusOK, "token refreshed successfully", payload)
}

func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserID(r.Context())
	if !ok {
		response.Write(w, http.StatusUnauthorized, "authentication required")
		return
	}

	sessionID, ok := middleware.SessionID(r.Context())
	if !ok {
		response.Write(w, http.StatusUnauthorized, "authentication required")
		return
	}

	clientID := r.Header.Get("X-Client-ID")
	if clientID == "" {
		response.Write(w, http.StatusBadRequest, "X-Client-ID header is required")
		return
	}

	if err := h.sessionService.Logout(r.Context(), sessionID); err != nil {
		log.Printf("Logout error: %v", err)
		response.Write(w, http.StatusInternalServerError, "logout failed")
		return
	}

	if err := h.clientDisconnecter.DisconnectClient(userID, clientID); err != nil {
		log.Printf("Logout websocket disconnect error: %v", err)
	}

	response.Write(w, http.StatusOK, "logout successful")
}
