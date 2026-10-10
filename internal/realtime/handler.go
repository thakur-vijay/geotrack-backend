package realtime

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/coder/websocket"
	"github.com/thakur-vijay/geotrack-backend/internal/middleware"
	"github.com/thakur-vijay/geotrack-backend/internal/modules/users"
)

type GroupRepository interface {
	GetMembers(ctx context.Context, groupID string) ([]users.User, error)
}

type Handler struct {
	manager         *Manager
	groupRepository GroupRepository
}

func NewHandler(
	manager *Manager,
	groupRepository GroupRepository,
) *Handler {
	return &Handler{
		manager:         manager,
		groupRepository: groupRepository,
	}
}

type IncomingEvent struct {
	Type string          `json:"type"`
	Data json.RawMessage `json:"data"`
}

type LocationUpdate struct {
	GroupID   string  `json:"group_id"`
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

type LocationUpdatedEvent struct {
	Type string `json:"type"`
	Data struct {
		GroupID   string  `json:"group_id"`
		UserID    string  `json:"user_id"`
		Latitude  float64 `json:"latitude"`
		Longitude float64 `json:"longitude"`
	} `json:"data"`
}

func (h *Handler) WebSocket(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserID(r.Context())
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	conn, err := websocket.Accept(w, r, nil)
	if err != nil {
		return
	}

	h.manager.Add(userID, conn)

	defer h.manager.Remove(userID, conn)
	defer conn.Close(websocket.StatusNormalClosure, "")

	for {
		_, message, err := conn.Read(r.Context())
		if err != nil {
			return
		}

		var incoming IncomingEvent
		if err := json.Unmarshal(message, &incoming); err != nil {
			continue
		}

		if incoming.Type != "location.updated" {
			continue
		}

		var location LocationUpdate
		if err := json.Unmarshal(incoming.Data, &location); err != nil {
			continue
		}

		if location.GroupID == "" ||
			location.Latitude < -90 ||
			location.Latitude > 90 ||
			location.Longitude < -180 ||
			location.Longitude > 180 {
			continue
		}

		members, err := h.groupRepository.GetMembers(
			r.Context(),
			location.GroupID,
		)
		if err != nil {
			continue
		}

		isMember := false
		memberIDs := make([]string, 0, len(members))

		for _, member := range members {
			memberIDs = append(memberIDs, member.ID)

			if member.ID == userID {
				isMember = true
			}
		}

		if !isMember {
			continue
		}

		event := LocationUpdatedEvent{
			Type: "location.updated",
		}

		event.Data.GroupID = location.GroupID
		event.Data.UserID = userID
		event.Data.Latitude = location.Latitude
		event.Data.Longitude = location.Longitude

		payload, err := json.Marshal(event)
		if err != nil {
			continue
		}

		h.manager.BroadcastToUsers(
			r.Context(),
			memberIDs,
			payload,
		)
	}
}
