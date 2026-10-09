package realtime

import "context"

type Publisher struct {
	manager *Manager
}

func NewPublisher(manager *Manager) *Publisher {
	return &Publisher{
		manager: manager,
	}
}

// Publish sends the event to the user's clients other than the publishing client.
// Exactly one non-empty client ID is required.
func (p *Publisher) Publish(
	ctx context.Context,
	userID string,
	event Event,
	clientIDs ...string,
) error {
	if len(clientIDs) != 1 || clientIDs[0] == "" {
		return ErrClientIDRequired
	}

	return p.manager.SendEvent(ctx, userID, clientIDs[0], event)
}
