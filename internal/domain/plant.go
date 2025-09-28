package domain

import (
	"time"
	"github.com/google/uuid"
)

// PlantStatus represents the health status of a plant
type PlantStatus string

const (
	PlantStatusHealthy   PlantStatus = "HEALTHY"
	PlantStatusAttention PlantStatus = "ATTENTION"
	PlantStatusCritical  PlantStatus = "CRITICAL"
)

// Plant represents a monitored plant in the system
type Plant struct {
	ID               uuid.UUID        `json:"id"`
	Name             string           `json:"name"`
	TypeID           *uuid.UUID       `json:"type_id,omitempty"`
	Type             *PlantType       `json:"type,omitempty"`
	MicrocontrollerID *uuid.UUID      `json:"microcontroller_id,omitempty"`
	Microcontroller  *Microcontroller `json:"microcontroller,omitempty"`
	Score            float64          `json:"score"` // Health score 0-100
	Change           string           `json:"change"` // Percentage change "+5.2%" or "-2.1%"
	Status           PlantStatus      `json:"status"`
	Temperature      *float64         `json:"temperature,omitempty"`
	Humidity         *float64         `json:"humidity,omitempty"` // Air humidity
	LightLevel       *float64         `json:"light_level,omitempty"`
	SoilHumidity     *float64         `json:"soil_humidity,omitempty"`
	LocationID       *uuid.UUID       `json:"location_id,omitempty"`
	Location         *Location        `json:"location,omitempty"`
	LocationName     *string          `json:"location_name,omitempty"` // Direct location string
	OwnerUserID      *uuid.UUID       `json:"owner_user_id,omitempty"`
	Owner            *User            `json:"owner,omitempty"`
	LastUpdate       time.Time        `json:"last_update"`
	Image            *string          `json:"image,omitempty"`
	ImageURL         *string          `json:"image_url,omitempty"` // From plant management service
	AlertsCount      int              `json:"alerts_count"`
	Thresholds       []*Threshold     `json:"thresholds,omitempty"`
	CreatedAt        time.Time        `json:"created_at"`
	UpdatedAt        time.Time        `json:"updated_at"`
}

// PlantType represents a type/category of plant with optimal conditions
type PlantType struct {
	ID                    uuid.UUID `json:"id"`
	Name                  string    `json:"name"` // e.g., "Tomate", "Lechuga"
	Description           *string   `json:"description,omitempty"`
	OptimalTemperature    Range     `json:"optimal_temperature"`
	OptimalHumidity       Range     `json:"optimal_humidity"`
	OptimalSoilHumidity   Range     `json:"optimal_soil_humidity"`
	OptimalLightLevel     Range     `json:"optimal_light_level"`
	Image                 *string   `json:"image,omitempty"`
	Plants                []*Plant  `json:"plants,omitempty"`
}

// Range represents a min-max range for optimal conditions
type Range struct {
	Min float64 `json:"min"`
	Max float64 `json:"max"`
}

// Threshold represents monitoring thresholds for a plant
type Threshold struct {
	ID          uuid.UUID   `json:"id"`
	PlantID     uuid.UUID   `json:"plant_id"`
	Plant       *Plant      `json:"plant,omitempty"`
	SensorType  SensorType  `json:"sensor_type"`
	MinValue    float64     `json:"min_value"`
	MaxValue    float64     `json:"max_value"`
	CriticalMin float64     `json:"critical_min"`
	CriticalMax float64     `json:"critical_max"`
	Unit        string      `json:"unit"`
}

// SensorType represents the type of sensor measurement
type SensorType string

const (
	SensorTypeTemperature SensorType = "TEMPERATURE"
	SensorTypeHumidity    SensorType = "HUMIDITY"
	SensorTypeSoil        SensorType = "SOIL"
	SensorTypeLight       SensorType = "LIGHT"
)