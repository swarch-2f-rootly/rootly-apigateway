package domain

import (
	"time"
	"github.com/google/uuid"
)

// SensorStatus represents the operational status of a sensor/microcontroller
type SensorStatus string

const (
	SensorStatusActive      SensorStatus = "ACTIVE"
	SensorStatusInactive    SensorStatus = "INACTIVE"
	SensorStatusMaintenance SensorStatus = "MAINTENANCE"
	SensorStatusError       SensorStatus = "ERROR"
)

// DeviceRole represents the permission level a user has on a device
type DeviceRole string

const (
	DeviceRoleViewer DeviceRole = "VIEWER"
	DeviceRoleEditor DeviceRole = "EDITOR"
	DeviceRoleOwner  DeviceRole = "OWNER"
)

// Microcontroller represents a physical IoT device for monitoring plants
type Microcontroller struct {
	ID               uuid.UUID                        `json:"id"`
	UniqueID         string                           `json:"unique_id"` // Physical device identifier (e.g., "ESP8266-001")
	Type             string                           `json:"type"`
	Location         *string                          `json:"location,omitempty"`
	Enabled          bool                             `json:"enabled"`
	PlantID          *uuid.UUID                       `json:"plant_id,omitempty"`
	Plant            *Plant                           `json:"plant,omitempty"`
	Status           SensorStatus                     `json:"status"`
	LastReading      *time.Time                       `json:"last_reading,omitempty"`
	IsActive         bool                             `json:"is_active"`
	BatteryLevel     *float64                         `json:"battery_level,omitempty"`
	SignalStrength   *float64                         `json:"signal_strength,omitempty"`
	UserAssociations []*UserMicrocontrollerAssociation `json:"user_associations,omitempty"`
	CreatedAt        time.Time                        `json:"created_at"`
}

// UserMicrocontrollerAssociation represents the relationship between a user and a microcontroller with permissions
type UserMicrocontrollerAssociation struct {
	UserID            uuid.UUID        `json:"user_id"`
	User              *User            `json:"user,omitempty"`
	MicrocontrollerID uuid.UUID        `json:"microcontroller_id"`
	Microcontroller   *Microcontroller `json:"microcontroller,omitempty"`
	Role              DeviceRole       `json:"role"`
	CreatedAt         time.Time        `json:"created_at"`
}

// Sensor represents a legacy sensor for backward compatibility
type Sensor struct {
	ID               string           `json:"id"` // Maps to microcontroller.unique_id
	PlantID          *uuid.UUID       `json:"plant_id,omitempty"`
	Plant            *Plant           `json:"plant,omitempty"`
	Microcontroller  *Microcontroller `json:"microcontroller,omitempty"`
	Status           SensorStatus     `json:"status"`
	LastReading      *time.Time       `json:"last_reading,omitempty"`
	Location         *Location        `json:"location,omitempty"`
	IsActive         bool             `json:"is_active"`
	BatteryLevel     *float64         `json:"battery_level,omitempty"`
	SignalStrength   *float64         `json:"signal_strength,omitempty"`
}

// RealTimeData represents current sensor measurements
type RealTimeData struct {
	PlantID           *uuid.UUID       `json:"plant_id,omitempty"`
	Plant             *Plant           `json:"plant,omitempty"`
	SensorID          string           `json:"sensor_id"`
	MicrocontrollerID *uuid.UUID       `json:"microcontroller_id,omitempty"`
	Microcontroller   *Microcontroller `json:"microcontroller,omitempty"`
	Timestamp         time.Time        `json:"timestamp"`
	Temperature       *float64         `json:"temperature,omitempty"`
	AirHumidity       *float64         `json:"air_humidity,omitempty"`
	SoilHumidity      *float64         `json:"soil_humidity,omitempty"`
	LightLevel        *float64         `json:"light_level,omitempty"`
}

// ChartData represents historical data points for charts
type ChartData struct {
	PlantID      uuid.UUID `json:"plant_id"`
	Time         time.Time `json:"time"`
	Temperature  *float64  `json:"temperature,omitempty"`
	Humidity     *float64  `json:"humidity,omitempty"` // Air humidity
	SoilHumidity *float64  `json:"soil_humidity,omitempty"`
	LightLevel   *float64  `json:"light_level,omitempty"`
}

// UserDevice represents a device associated with a user (for API responses)
type UserDevice struct {
	ID              uuid.UUID        `json:"id"`
	UniqueID        string           `json:"unique_id"`
	Type            string           `json:"type"`
	Location        *string          `json:"location,omitempty"`
	Enabled         bool             `json:"enabled"`
	Role            DeviceRole       `json:"role"`
	Plant           *Plant           `json:"plant,omitempty"`
	Microcontroller *Microcontroller `json:"microcontroller"`
}

// UserDeviceList represents a paginated list of user devices
type UserDeviceList struct {
	Devices     []*UserDevice `json:"devices"`
	TotalCount  int           `json:"total_count"`
	HasNextPage bool          `json:"has_next_page"`
}