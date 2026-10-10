package realtime

import (
	"context"
	"encoding/json"
)

type Publisher struct {
	manager *Manager
}

func NewPublisher(manager *Manager) *Publisher {
	return &Publisher{
		manager: manager,
	}
}

func (p *Publisher) Publish(
	ctx context.Context,
	userID string,
	event Event,
) error {
	message, err := json.Marshal(event)
	if err != nil {
		return err
	}

	p.manager.Broadcast(ctx, message)

	return nil
}
