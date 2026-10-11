package groups

import (
	"time"

	"github.com/thakur-vijay/geotrack-backend/internal/modules/users"
)

type Group struct {
	ID        string       `json:"id"`
	Name      string       `json:"name"`
	Members   []users.User `json:"members"`
	CreatedBy string       `json:"created_by"`
	CreatedAt time.Time    `json:"created_at"`
	UpdatedAt time.Time    `json:"updated_at"`
}

type CreateGroupDTO struct {
	Name string `json:"name" validate:"required"`
}
