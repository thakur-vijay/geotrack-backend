package realtime

import (
	"context"
	"encoding/json"
	"math"
	"net/http"

	"github.com/coder/websocket"
	"github.com/thakur-vijay/geotrack-backend/internal/middleware"
	"github.com/thakur-vijay/geotrack-backend/internal/modules/regions"
	"github.com/thakur-vijay/geotrack-backend/internal/modules/users"
)

type SafeRegionRepository interface {
	FetchByGroupID(
		ctx context.Context,
		groupID string,
		userID string,
	) ([]regions.SafeRegion, error)
}

type GroupRepository interface {
	GetMembers(ctx context.Context, groupID string) ([]users.User, error)
}

type Handler struct {
	manager              *Manager
	groupRepository      GroupRepository
	safeRegionRepository SafeRegionRepository
}

func NewHandler(
	manager *Manager,
	groupRepository GroupRepository,
	safeRegionRepository SafeRegionRepository,
) *Handler {
	return &Handler{
		manager:              manager,
		groupRepository:      groupRepository,
		safeRegionRepository: safeRegionRepository,
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

type UserWentOutsideEvent struct {
	Type string `json:"type"`
	Data struct {
		GroupID   string             `json:"group_id"`
		User      users.User         `json:"user"`
		Latitude  float64            `json:"latitude"`
		Longitude float64            `json:"longitude"`
		Region    regions.SafeRegion `json:"region"`
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

		safeRegions, err := h.safeRegionRepository.FetchByGroupID(
			r.Context(),
			location.GroupID,
			userID,
		)
		if err != nil || len(safeRegions) == 0 {
			continue
		}

		region := safeRegions[0]

		distance := distanceMeters(
			location.Latitude,
			location.Longitude,
			region.Latitude,
			region.Longitude,
		)

		if distance > float64(region.RadiusMeters) {
			var outsideEvent UserWentOutsideEvent
			outsideEvent.Type = "user_went_outside"
			outsideEvent.Data.GroupID = location.GroupID
			outsideEvent.Data.User = users.User{
				ID: userID,
			}
			outsideEvent.Data.Latitude = location.Latitude
			outsideEvent.Data.Longitude = location.Longitude
			outsideEvent.Data.Region = region

			outsidePayload, err := json.Marshal(outsideEvent)
			if err == nil {
				h.manager.BroadcastToUsers(
					r.Context(),
					memberIDs,
					outsidePayload,
				)
			}
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

func distanceMeters(
	lat1, lon1, lat2, lon2 float64,
) float64 {
	const earthRadius = 6371000.0

	toRadians := func(degrees float64) float64 {
		return degrees * 3.141592653589793 / 180
	}

	dLat := toRadians(lat2 - lat1)
	dLon := toRadians(lon2 - lon1)

	a := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Cos(toRadians(lat1))*
			math.Cos(toRadians(lat2))*
			math.Sin(dLon/2)*math.Sin(dLon/2)

	return earthRadius * 2 * math.Atan2(
		math.Sqrt(a),
		math.Sqrt(1-a),
	)
}
