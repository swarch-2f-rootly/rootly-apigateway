package domain

import "time"

// Plant represents a monitored plant entity
type Plant struct {
	ID            string    `json:"id"`
	Name          string    `json:"name"`
	Species       string    `json:"species"`
	Description   *string   `json:"description,omitempty"`
	UserID        string    `json:"user_id"`
	PhotoFilename *string   `json:"photo_filename,omitempty"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// CreatePlantRequest represents the input for creating a plant
type CreatePlantRequest struct {
	Name          string  `json:"name" validate:"required"`
	Species       string  `json:"species" validate:"required"`
	Description   *string `json:"description,omitempty"`
	UserID        string  `json:"user_id" validate:"required"`
	PhotoFilename *string `json:"photo_filename,omitempty"`
}

// UpdatePlantRequest represents the input for updating a plant
type UpdatePlantRequest struct {
	Name          *string `json:"name,omitempty"`
	Species       *string `json:"species,omitempty"`
	Description   *string `json:"description,omitempty"`
	PhotoFilename *string `json:"photo_filename,omitempty"`
}

// PlantFilter represents filters for plant queries
type PlantFilter struct {
	UserID       *string `json:"user_id,omitempty"`
	Species      *string `json:"species,omitempty"`
	NameContains *string `json:"name_contains,omitempty"`
	Limit        *int    `json:"limit,omitempty"`
	Offset       *int    `json:"offset,omitempty"`
}

// PlantsResponse represents the response for plant list queries
type PlantsResponse struct {
	Plants []Plant `json:"plants"`
	Total  int     `json:"total"`
	Page   int     `json:"page"`
	Limit  int     `json:"limit"`
}

// PlantResponse represents the response for single plant queries
type PlantResponse struct {
	Plant *Plant `json:"plant,omitempty"`
}

// PlantMutationResponse represents the response for plant mutations
type PlantMutationResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Plant   *Plant `json:"plant,omitempty"`
}

// PlantDeleteResponse represents the response for plant deletion
type PlantDeleteResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}
