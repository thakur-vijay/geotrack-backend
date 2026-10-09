package realtime

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"sync"

	"github.com/coder/websocket"
)

// ErrClientIDRequired is returned when a publish operation has no client ID.
var ErrClientIDRequired = errors.New("client ID is required")

type Manager struct {
	mu          sync.RWMutex
	connections map[string]map[*websocket.Conn]string
}

func NewManager() *Manager {
	log.Println("[Realtime] Creating WebSocket manager")

	return &Manager{
		connections: make(map[string]map[*websocket.Conn]string),
	}
}

func (m *Manager) Add(userID, clientID string, conn *websocket.Conn) {
	log.Printf("[Realtime] Adding connection for user: %s", userID)

	m.mu.Lock()
	defer m.mu.Unlock()

	if m.connections[userID] == nil {
		log.Printf("[Realtime] Creating connection pool for user: %s", userID)

		m.connections[userID] = make(map[*websocket.Conn]string)
	}

	m.connections[userID][conn] = clientID

	log.Printf(
		"[Realtime] Connection added | user=%s | connections=%d",
		userID,
		len(m.connections[userID]),
	)
}

func (m *Manager) Remove(userID string, conn *websocket.Conn) {
	log.Printf("[Realtime] Removing connection for user: %s", userID)

	m.mu.Lock()
	defer m.mu.Unlock()

	connections, ok := m.connections[userID]
	if !ok {
		log.Printf("[Realtime] No connection pool found for user: %s", userID)
		return
	}

	delete(connections, conn)

	log.Printf(
		"[Realtime] Connection removed | user=%s | remaining=%d",
		userID,
		len(connections),
	)

	if len(connections) == 0 {
		delete(m.connections, userID)

		log.Printf("[Realtime] Removed empty connection pool for user: %s", userID)
	}
}

func (m *Manager) DisconnectClient(userID, clientID string) error {
	if clientID == "" {
		return ErrClientIDRequired
	}

	m.mu.Lock()
	connections := m.connections[userID]
	clientConnections := make([]*websocket.Conn, 0)
	for conn, connectionClientID := range connections {
		if connectionClientID == clientID {
			clientConnections = append(clientConnections, conn)
			delete(connections, conn)
		}
	}
	if len(connections) == 0 {
		delete(m.connections, userID)
	}
	m.mu.Unlock()

	var disconnectErr error
	for _, conn := range clientConnections {
		if err := conn.CloseNow(); err != nil {
			log.Printf(
				"[Realtime] Failed to disconnect client | user=%s | client=%s | error=%v",
				userID,
				clientID,
				err,
			)
			disconnectErr = errors.Join(disconnectErr, err)
		}
	}

	return disconnectErr
}

func (m *Manager) Send(
	ctx context.Context,
	userID string,
	clientID string,
	message []byte,
) error {
	if clientID == "" {
		return ErrClientIDRequired
	}

	log.Printf(
		"[Realtime] Sending message | user=%s | payload_size=%d bytes",
		userID,
		len(message),
	)

	m.mu.RLock()

	connections := make([]*websocket.Conn, 0, len(m.connections[userID]))

	for conn, connectionClientID := range m.connections[userID] {
		if connectionClientID == clientID {
			continue
		}

		connections = append(connections, conn)
	}

	connectionCount := len(connections)

	m.mu.RUnlock()

	log.Printf(
		"[Realtime] Broadcasting message | user=%s | connections=%d",
		userID,
		connectionCount,
	)

	var sendErr error

	for _, conn := range connections {
		if err := conn.Write(ctx, websocket.MessageText, message); err != nil {
			log.Printf(
				"[Realtime] Failed to send message | user=%s | error=%v",
				userID,
				err,
			)

			// Remove stale/dead connection.
			m.Remove(userID, conn)

			// Don't stop broadcasting to other devices.
			sendErr = err
			continue
		}

		log.Printf(
			"[Realtime] Message sent successfully | user=%s",
			userID,
		)
	}

	return sendErr
}

func (m *Manager) SendEvent(
	ctx context.Context,
	userID string,
	clientID string,
	event Event,
) error {
	if clientID == "" {
		return ErrClientIDRequired
	}

	log.Printf(
		"[Realtime] Sending event | user=%s | event=%+v",
		userID,
		event,
	)

	message, err := json.Marshal(event)
	if err != nil {
		log.Printf(
			"[Realtime] Failed to encode event | user=%s | error=%v",
			userID,
			err,
		)

		return err
	}

	log.Printf(
		"[Realtime] Event encoded | user=%s | payload_size=%d bytes",
		userID,
		len(message),
	)

	return m.Send(ctx, userID, clientID, message)
}
