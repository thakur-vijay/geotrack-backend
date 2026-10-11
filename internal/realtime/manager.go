package realtime

import (
	"context"
	"log"
	"sync"

	"github.com/coder/websocket"
)

type Manager struct {
	mu          sync.RWMutex
	connections map[string]*websocket.Conn
}

func NewManager() *Manager {
	return &Manager{
		connections: make(map[string]*websocket.Conn),
	}
}

func (m *Manager) Add(userID string, conn *websocket.Conn) {
	m.mu.Lock()
	oldConn := m.connections[userID]
	m.connections[userID] = conn
	m.mu.Unlock()

	if oldConn != nil {
		_ = oldConn.CloseNow()
	}
}

func (m *Manager) Remove(userID string, conn *websocket.Conn) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.connections[userID] == conn {
		delete(m.connections, userID)
	}
}

func (m *Manager) Broadcast(ctx context.Context, message []byte) {
	m.mu.RLock()
	connections := make(map[string]*websocket.Conn, len(m.connections))

	for userID, conn := range m.connections {
		connections[userID] = conn
	}
	m.mu.RUnlock()

	for userID, conn := range connections {
		if err := conn.Write(ctx, websocket.MessageText, message); err != nil {
			log.Printf("[Realtime] Broadcast failed | user=%s | error=%v", userID, err)
			m.Remove(userID, conn)
		}
	}
}

func (m *Manager) BroadcastToUsers(
	ctx context.Context,
	userIDs []string,
	message []byte,
) {
	m.mu.RLock()
	connections := make(map[string]*websocket.Conn, len(userIDs))

	for _, userID := range userIDs {
		if conn, exists := m.connections[userID]; exists {
			connections[userID] = conn
		}
	}

	m.mu.RUnlock()

	for userID, conn := range connections {
		if err := conn.Write(ctx, websocket.MessageText, message); err != nil {
			log.Printf(
				"[Realtime] Broadcast failed | user=%s | error=%v",
				userID,
				err,
			)
			m.Remove(userID, conn)
		}
	}
}

func (m *Manager) DisconnectUser(userID string) {
	m.mu.Lock()
	conn := m.connections[userID]
	delete(m.connections, userID)
	m.mu.Unlock()

	if conn != nil {
		_ = conn.Close(websocket.StatusNormalClosure, "user logged out")
	}
}
