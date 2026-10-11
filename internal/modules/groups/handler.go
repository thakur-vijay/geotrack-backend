package groups

import (
	"database/sql"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/thakur-vijay/geotrack-backend/internal/middleware"
	"github.com/thakur-vijay/geotrack-backend/internal/response"
)

type OnlineChecker interface {
	IsOnline(userID string) bool
}

type Handler struct {
	repository    *Repository
	onlineChecker OnlineChecker
}

func NewHandler(repository *Repository, onlineChecker OnlineChecker) *Handler {
	return &Handler{
		repository:    repository,
		onlineChecker: onlineChecker,
	}
}

func (h *Handler) setMembersOnlineStatus(group *Group) {
	for i := range group.Members {
		group.Members[i].IsOnline = h.onlineChecker.IsOnline(group.Members[i].ID)
	}
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	var dto CreateGroupDTO

	if err := response.DecodeJSON(r, &dto); err != nil {
		response.Write(w, http.StatusBadRequest, "invalid request body")
		return
	}

	dto.Name = strings.TrimSpace(dto.Name)
	if dto.Name == "" {
		response.Write(w, http.StatusBadRequest, "group name is required")
		return
	}

	userID, _ := middleware.UserID(r.Context())
	if userID == "" {
		response.Write(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	now := time.Now().UTC()

	group := Group{
		ID:        uuid.NewString(),
		Name:      dto.Name,
		CreatedBy: userID,
		CreatedAt: now,
		UpdatedAt: now,
	}

	if err := h.repository.Create(r.Context(), group, userID); err != nil {
		log.Printf("Create group error: %v", err)
		response.Write(w, http.StatusInternalServerError, "group could not be created")
		return
	}

	createdGroup, err := h.repository.GetByID(r.Context(), group.ID)
	if err != nil {
		log.Printf("Get created group error: %v", err)
		response.Write(w, http.StatusInternalServerError, "group was created but could not be retrieved")
		return
	}

	response.WriteItem(w, http.StatusCreated, "group created successfully", createdGroup)
}

func (h *Handler) GetByID(w http.ResponseWriter, r *http.Request) {
	groupID := r.PathValue("id")

	group, err := h.repository.GetByID(r.Context(), groupID)
	if err != nil {
		if err == sql.ErrNoRows {
			response.Write(w, http.StatusNotFound, "group was not found")
			return
		}

		log.Printf("GetByID group error: %v", err)
		response.Write(w, http.StatusInternalServerError, "group could not be retrieved")
		return
	}

	userID, _ := middleware.UserID(r.Context())
	isMember := false

	for _, member := range group.Members {
		if member.ID == userID {
			isMember = true
			break
		}
	}

	if !isMember {
		response.Write(w, http.StatusForbidden, "you are not a member of this group")
		return
	}

	h.setMembersOnlineStatus(&group)
	response.WriteItem(w, http.StatusOK, "group retrieved successfully", group)
}

func (h *Handler) GetMyGroups(w http.ResponseWriter, r *http.Request) {
	userID, _ := middleware.UserID(r.Context())
	if userID == "" {
		response.Write(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	groups, err := h.repository.GetByUserID(r.Context(), userID)
	if err != nil {
		log.Printf("GetMyGroups error: %v", err)
		response.Write(w, http.StatusInternalServerError, "groups could not be retrieved")
		return
	}

	for i := range groups {
		h.setMembersOnlineStatus(&groups[i])
	}
	response.WriteItem(w, http.StatusOK, "groups retrieved successfully", groups)
}

func (h *Handler) AddMember(w http.ResponseWriter, r *http.Request) {
	groupID := r.PathValue("id")

	var dto struct {
		UserID string `json:"user_id"`
	}

	if err := response.DecodeJSON(r, &dto); err != nil {
		response.Write(w, http.StatusBadRequest, "invalid request body")
		return
	}

	dto.UserID = strings.TrimSpace(dto.UserID)
	if dto.UserID == "" {
		response.Write(w, http.StatusBadRequest, "user_id is required")
		return
	}

	if _, err := uuid.Parse(groupID); err != nil {
		response.Write(w, http.StatusBadRequest, "invalid group id")
		return
	}

	if _, err := uuid.Parse(dto.UserID); err != nil {
		response.Write(w, http.StatusBadRequest, "invalid user id")
		return
	}

	currentUserID, _ := middleware.UserID(r.Context())

	group, err := h.repository.GetByID(r.Context(), groupID)
	if err != nil {
		if err == sql.ErrNoRows {
			response.Write(w, http.StatusNotFound, "group was not found")
			return
		}

		log.Printf("AddMember group lookup error: %v", err)
		response.Write(w, http.StatusInternalServerError, "group could not be retrieved")
		return
	}

	isMember := false
	for _, member := range group.Members {
		if member.ID == currentUserID {
			isMember = true
			break
		}
	}

	if !isMember {
		response.Write(w, http.StatusForbidden, "you are not a member of this group")
		return
	}

	if err := h.repository.AddMember(r.Context(), groupID, dto.UserID); err != nil {
		log.Printf("AddMember error: %v", err)
		response.Write(w, http.StatusInternalServerError, "member could not be added")
		return
	}

	updatedGroup, err := h.repository.GetByID(r.Context(), groupID)
	if err != nil {
		log.Printf("Get updated group error: %v", err)
		response.Write(w, http.StatusInternalServerError, "member was added but group could not be retrieved")
		return
	}

	response.WriteItem(w, http.StatusOK, "member added successfully", updatedGroup)
}

func (h *Handler) RemoveMember(w http.ResponseWriter, r *http.Request) {
	groupID := r.PathValue("id")
	memberID := r.PathValue("userID")

	currentUserID, _ := middleware.UserID(r.Context())

	if _, err := uuid.Parse(groupID); err != nil {
		response.Write(w, http.StatusBadRequest, "invalid group id")
		return
	}

	if _, err := uuid.Parse(memberID); err != nil {
		response.Write(w, http.StatusBadRequest, "invalid user id")
		return
	}

	group, err := h.repository.GetByID(r.Context(), groupID)
	if err != nil {
		if err == sql.ErrNoRows {
			response.Write(w, http.StatusNotFound, "group was not found")
			return
		}

		log.Printf("RemoveMember group lookup error: %v", err)
		response.Write(w, http.StatusInternalServerError, "group could not be retrieved")
		return
	}

	isMember := false
	for _, member := range group.Members {
		if member.ID == currentUserID {
			isMember = true
			break
		}
	}

	if !isMember {
		response.Write(w, http.StatusForbidden, "you are not a member of this group")
		return
	}

	if err := h.repository.RemoveMember(r.Context(), groupID, memberID); err != nil {
		if err == sql.ErrNoRows {
			response.Write(w, http.StatusNotFound, "member was not found in this group")
			return
		}

		log.Printf("RemoveMember error: %v", err)
		response.Write(w, http.StatusInternalServerError, "member could not be removed")
		return
	}

	response.Write(w, http.StatusOK, "member removed successfully")
}

func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	groupID := r.PathValue("id")
	currentUserID, _ := middleware.UserID(r.Context())

	if _, err := uuid.Parse(groupID); err != nil {
		response.Write(w, http.StatusBadRequest, "invalid group id")
		return
	}

	group, err := h.repository.GetByID(r.Context(), groupID)
	if err != nil {
		if err == sql.ErrNoRows {
			response.Write(w, http.StatusNotFound, "group was not found")
			return
		}

		log.Printf("Delete group lookup error: %v", err)
		response.Write(w, http.StatusInternalServerError, "group could not be retrieved")
		return
	}

	isMember := false
	for _, member := range group.Members {
		if member.ID == currentUserID {
			isMember = true
			break
		}
	}

	if !isMember {
		response.Write(w, http.StatusForbidden, "you are not a member of this group")
		return
	}

	if err := h.repository.Delete(r.Context(), groupID); err != nil {
		if err == sql.ErrNoRows {
			response.Write(w, http.StatusNotFound, "group was not found")
			return
		}

		log.Printf("Delete group error: %v", err)
		response.Write(w, http.StatusInternalServerError, "group could not be deleted")
		return
	}

	response.Write(w, http.StatusOK, "group deleted successfully")
}
