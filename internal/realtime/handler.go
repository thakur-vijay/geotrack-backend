package realtime

import (
	"net/http"

	"github.com/coder/websocket"
	"github.com/thakur-vijay/geotrack-backend/internal/middleware"
	"github.com/thakur-vijay/geotrack-backend/internal/response"
)

const clientIDHeader = "X-Client-ID"

type Handler struct {
	manager *Manager
}

func NewHandler(manager *Manager) *Handler {
	return &Handler{
		manager: manager,
	}
}

func (h *Handler) WebSocket(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserID(r.Context())
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	clientID := r.Header.Get(clientIDHeader)
	if clientID == "" {
		response.Write(w, http.StatusBadRequest, "X-Client-ID header is required")
		return
	}

	conn, err := websocket.Accept(w, r, nil)
	if err != nil {
		return
	}

	h.manager.Add(userID, clientID, conn)

	defer h.manager.Remove(userID, conn)
	defer conn.Close(websocket.StatusNormalClosure, "")

	for {
		_, _, err := conn.Read(r.Context())
		if err != nil {
			return
		}
	}
}
