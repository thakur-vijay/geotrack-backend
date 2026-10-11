package realtime

import (
	"context"
	"encoding/json"
	"log"
	"math"
	"net/http"
	"sync"

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

	mu             sync.Mutex
	outsideRegions map[string]bool
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
		outsideRegions:       make(map[string]bool),
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
	log.Println("[WS] WebSocket connection request received")

	userID, ok := middleware.UserID(r.Context())
	if !ok {
		log.Println("[WS] ERROR: User ID missing from context")
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	log.Printf("[WS] Authenticated user: %s", userID)

	conn, err := websocket.Accept(w, r, nil)
	if err != nil {
		log.Printf("[WS] ERROR: WebSocket accept failed: %v", err)
		return
	}

	log.Printf("[WS] Connection accepted for user: %s", userID)

	h.manager.Add(userID, conn)
	log.Printf("[WS] Connection registered for user: %s", userID)

	defer func() {
		log.Printf("[WS] Removing connection for user: %s", userID)
		h.manager.Remove(userID, conn)
	}()

	defer conn.Close(websocket.StatusNormalClosure, "")

	for {
		_, message, err := conn.Read(r.Context())
		if err != nil {
			log.Printf("[WS] Read failed for user %s: %v", userID, err)
			return
		}

		log.Printf("[WS] RAW MESSAGE from user %s: %s", userID, string(message))

		var incoming IncomingEvent
		if err := json.Unmarshal(message, &incoming); err != nil {
			log.Printf("[WS] ERROR: Incoming event JSON decode failed: %v", err)
			continue
		}

		log.Printf("[WS] Incoming event type: %q", incoming.Type)

		if incoming.Type != "location.updated" {
			log.Printf("[WS] Ignoring unsupported event type: %q", incoming.Type)
			continue
		}

		var location LocationUpdate
		if err := json.Unmarshal(incoming.Data, &location); err != nil {
			log.Printf("[WS] ERROR: Location payload decode failed: %v", err)
			continue
		}

		log.Printf(
			"[WS] LOCATION RECEIVED: user=%s group=%s lat=%.9f lon=%.9f",
			userID,
			location.GroupID,
			location.Latitude,
			location.Longitude,
		)

		if location.GroupID == "" ||
			location.Latitude < -90 ||
			location.Latitude > 90 ||
			location.Longitude < -180 ||
			location.Longitude > 180 {
			log.Printf(
				"[WS] ERROR: Invalid location payload: group=%q lat=%f lon=%f",
				location.GroupID,
				location.Latitude,
				location.Longitude,
			)
			continue
		}

		members, err := h.groupRepository.GetMembers(
			r.Context(),
			location.GroupID,
		)
		if err != nil {
			log.Printf(
				"[WS] ERROR: GetMembers failed: group=%s err=%v",
				location.GroupID,
				err,
			)
			continue
		}

		log.Printf(
			"[WS] Group members fetched: group=%s count=%d",
			location.GroupID,
			len(members),
		)

		isMember := false
		memberIDs := make([]string, 0, len(members))

		for _, member := range members {
			memberIDs = append(memberIDs, member.ID)

			log.Printf(
				"[WS] Group member: id=%s name=%q",
				member.ID,
				member.FullName,
			)

			if member.ID == userID {
				isMember = true
			}
		}

		if !isMember {
			log.Printf(
				"[WS] REJECTED: user %s is not a member of group %s",
				userID,
				location.GroupID,
			)
			continue
		}

		log.Printf("[WS] User membership verified: user=%s", userID)

		safeRegions, err := h.safeRegionRepository.FetchByGroupID(
			r.Context(),
			location.GroupID,
			userID,
		)
		if err != nil {
			log.Printf(
				"[WS] ERROR: FetchByGroupID failed: group=%s user=%s err=%v",
				location.GroupID,
				userID,
				err,
			)
			continue
		}

		log.Printf(
			"[WS] Safe regions fetched: group=%s user=%s count=%d",
			location.GroupID,
			userID,
			len(safeRegions),
		)

		if len(safeRegions) == 0 {
			log.Printf(
				"[WS] SKIPPED: No safe regions found for group=%s user=%s",
				location.GroupID,
				userID,
			)
			continue
		}

		for i, safeRegion := range safeRegions {
			log.Printf(
				"[WS] SAFE REGION [%d]: id=%s lat=%.9f lon=%.9f radius=%dm",
				i,
				safeRegion.ID,
				safeRegion.Latitude,
				safeRegion.Longitude,
				safeRegion.RadiusMeters,
			)
		}

		region := safeRegions[0]

		distance := distanceMeters(
			location.Latitude,
			location.Longitude,
			region.Latitude,
			region.Longitude,
		)

		isOutside := distance > float64(region.RadiusMeters)
		stateKey := userID + ":" + location.GroupID + ":" + region.ID

		log.Printf(
			"[WS] DISTANCE CHECK: user=%s region=%s distance=%.2fm radius=%dm isOutside=%t stateKey=%s",
			userID,
			region.ID,
			distance,
			region.RadiusMeters,
			isOutside,
			stateKey,
		)

		h.mu.Lock()
		wasOutside := h.outsideRegions[stateKey]
		h.outsideRegions[stateKey] = isOutside
		h.mu.Unlock()

		log.Printf(
			"[WS] OUTSIDE STATE: wasOutside=%t isOutside=%t",
			wasOutside,
			isOutside,
		)

		if isOutside && !wasOutside {
			log.Printf(
				"[WS] EXIT DETECTED: user=%s group=%s region=%s",
				userID,
				location.GroupID,
				region.ID,
			)

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
			if err != nil {
				log.Printf("[WS] ERROR: Outside event marshal failed: %v", err)
			} else {
				log.Printf(
					"[WS] OUTSIDE EVENT PAYLOAD: %s",
					string(outsidePayload),
				)

				log.Printf(
					"[WS] Broadcasting outside event to %d group members",
					len(memberIDs),
				)

				h.manager.BroadcastToUsers(
					r.Context(),
					memberIDs,
					outsidePayload,
				)

				log.Printf("[WS] Outside event broadcast call completed")
			}
		} else if isOutside && wasOutside {
			log.Printf(
				"[WS] No new exit event: user was already outside region=%s",
				region.ID,
			)
		} else {
			log.Printf(
				"[WS] User is inside safe region=%s",
				region.ID,
			)
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
			log.Printf("[WS] ERROR: Location event marshal failed: %v", err)
			continue
		}

		log.Printf(
			"[WS] Broadcasting location.updated to %d group members",
			len(memberIDs),
		)

		h.manager.BroadcastToUsers(
			r.Context(),
			memberIDs,
			payload,
		)

		log.Printf("[WS] location.updated broadcast call completed")
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
