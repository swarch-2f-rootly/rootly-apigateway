package domain

import (
	"time"
	"github.com/google/uuid"
)

// AlertType represents the type of alert
type AlertType string

const (
	AlertTypeTemperature   AlertType = "TEMPERATURE"
	AlertTypeHumidity      AlertType = "HUMIDITY"
	AlertTypeSoil          AlertType = "SOIL"
	AlertTypeLight         AlertType = "LIGHT"
	AlertTypeBattery       AlertType = "BATTERY"
	AlertTypeConnectivity  AlertType = "CONNECTIVITY"
)

// AlertPriority represents the priority level of an alert
type AlertPriority string

const (
	AlertPriorityLow      AlertPriority = "LOW"
	AlertPriorityMedium   AlertPriority = "MEDIUM"
	AlertPriorityHigh     AlertPriority = "HIGH"
	AlertPriorityCritical AlertPriority = "CRITICAL"
)

// Alert represents a monitoring alert
type Alert struct {
	ID         uuid.UUID     `json:"id"`
	PlantID    uuid.UUID     `json:"plant_id"`
	Plant      *Plant        `json:"plant,omitempty"`
	Type       AlertType     `json:"type"`
	Priority   AlertPriority `json:"priority"`
	Message    string        `json:"message"`
	Value      string        `json:"value"`
	ThresholdID *uuid.UUID   `json:"threshold_id,omitempty"`
	Threshold  *Threshold    `json:"threshold,omitempty"`
	Timestamp  time.Time     `json:"timestamp"`
	IsRead     bool          `json:"is_read"`
	ResolvedAt *time.Time    `json:"resolved_at,omitempty"`
	CreatedAt  time.Time     `json:"created_at"`
}

// Location represents a physical location
type Location struct {
	ID              uuid.UUID        `json:"id"`
	Name            string           `json:"name"`
	Description     *string          `json:"description,omitempty"`
	Coordinates     *Coordinates     `json:"coordinates,omitempty"`
	Plants          []*Plant         `json:"plants,omitempty"`
	Sensors         []*Sensor        `json:"sensors,omitempty"`
	Microcontrollers []*Microcontroller `json:"microcontrollers,omitempty"`
}

// Coordinates represents geographic coordinates
type Coordinates struct {
	Lat float64 `json:"lat"`
	Lng float64 `json:"lng"`
}

// GlobalStats represents system-wide statistics
type GlobalStats struct {
	TotalPlants             int       `json:"total_plants"`
	HealthyPlants           int       `json:"healthy_plants"`
	AlertPlants             int       `json:"alert_plants"`
	CriticalPlants          int       `json:"critical_plants"`
	AverageTemperature      float64   `json:"average_temperature"`
	AverageHumidity         float64   `json:"average_humidity"`
	ActiveSensors           int       `json:"active_sensors"`
	ActiveMicrocontrollers  int       `json:"active_microcontrollers"`
	TotalMicrocontrollers   int       `json:"total_microcontrollers"`
	EnabledMicrocontrollers int       `json:"enabled_microcontrollers"`
	Uptime                  float64   `json:"uptime"` // Percentage uptime
	LastUpdated             time.Time `json:"last_updated"`
}

// HealthStatus represents a health check response
type HealthStatus struct {
	Status    string     `json:"status"`
	Service   string     `json:"service"`
	Timestamp time.Time  `json:"timestamp"`
	Details   *string    `json:"details,omitempty"`
}

// DeviceOperationResult represents the result of a device operation
type DeviceOperationResult struct {
	ID      uuid.UUID        `json:"id"`
	Success bool             `json:"success"`
	Message string           `json:"message"`
	Device  *Microcontroller `json:"device,omitempty"`
}