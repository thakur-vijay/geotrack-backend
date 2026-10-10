package regions

import "time"

type SafeRegion struct {
	ID           string    `json:"id"`
	UserID       string    `json:"user_id"`
	GroupID      string    `json:"group_id"`
	Name         string    `json:"name"`
	Latitude     float64   `json:"latitude"`
	Longitude    float64   `json:"longitude"`
	RadiusMeters int       `json:"radius_meters"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type CreateRegionDTO struct {
	GroupID      string  `json:"group_id" validate:"required,uuid"`
	Name         string  `json:"name" validate:"required,notblank"`
	Latitude     float64 `json:"latitude" validate:"gte=-90,lte=90"`
	Longitude    float64 `json:"longitude" validate:"gte=-180,lte=180"`
	RadiusMeters int     `json:"radius_meters" validate:"required,gt=0"`
}

type UpdateRegionDTO struct {
	GroupID      *string  `json:"group_id" validate:"omitempty,uuid"`
	Name         *string  `json:"name" validate:"omitempty,notblank"`
	Latitude     *float64 `json:"latitude" validate:"omitempty,gte=-90,lte=90"`
	Longitude    *float64 `json:"longitude" validate:"omitempty,gte=-180,lte=180"`
	RadiusMeters *int     `json:"radius_meters" validate:"omitempty,gt=0"`
}
