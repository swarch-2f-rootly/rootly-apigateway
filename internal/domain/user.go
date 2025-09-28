package domain

import (
	"time"
	"github.com/google/uuid"
)

// User represents a user in the system
type User struct {
	ID                  uuid.UUID                        `json:"id"`
	Name                string                           `json:"name"` // Full name derived from first_name + last_name
	Email               string                           `json:"email"`
	FirstName           string                           `json:"first_name"`
	LastName            string                           `json:"last_name"`
	PasswordHash        string                           `json:"password_hash"` // Not exposed in GraphQL
	ProfilePhotoURL     *string                          `json:"profile_photo_url,omitempty"`
	RoleID              uuid.UUID                        `json:"role_id"`
	Role                *UserRole                        `json:"role,omitempty"`
	Permissions         []*Permission                    `json:"permissions,omitempty"`
	IsActive            bool                             `json:"is_active"`
	DeviceAssociations  []*UserMicrocontrollerAssociation `json:"device_associations,omitempty"`
	OwnedPlants         []*Plant                         `json:"owned_plants,omitempty"`
	Notifications       []*Notification                  `json:"notifications,omitempty"`
	Sessions            []*Session                       `json:"sessions,omitempty"`
	CreatedAt           time.Time                        `json:"created_at"`
	LastLogin           *time.Time                       `json:"last_login,omitempty"`
}

// UserRole represents a role with permissions
type UserRole struct {
	ID          uuid.UUID     `json:"id"`
	Name        string        `json:"name"`
	Description *string       `json:"description,omitempty"`
	Permissions []*Permission `json:"permissions,omitempty"`
	Users       []*User       `json:"users,omitempty"`
}

// Permission represents a system permission
type Permission struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	Description *string   `json:"description,omitempty"`
	Resource    string    `json:"resource"`
	Action      string    `json:"action"`
}

// Session represents a user session
type Session struct {
	ID           uuid.UUID  `json:"id"`
	UserID       uuid.UUID  `json:"user_id"`
	User         *User      `json:"user,omitempty"`
	Token        string     `json:"token"`
	RefreshToken *string    `json:"refresh_token,omitempty"`
	ExpiresAt    time.Time  `json:"expires_at"`
	DeviceInfo   *string    `json:"device_info,omitempty"`
	LastActivity time.Time  `json:"last_activity"`
	IsActive     bool       `json:"is_active"`
	CreatedAt    time.Time  `json:"created_at"`
}

// NotificationType represents the type of notification
type NotificationType string

const (
	NotificationTypeAlert   NotificationType = "ALERT"
	NotificationTypeInfo    NotificationType = "INFO"
	NotificationTypeWarning NotificationType = "WARNING"
	NotificationTypeSuccess NotificationType = "SUCCESS"
)

// Notification represents a user notification
type Notification struct {
	ID        uuid.UUID        `json:"id"`
	UserID    uuid.UUID        `json:"user_id"`
	User      *User            `json:"user,omitempty"`
	Type      NotificationType `json:"type"`
	Title     string           `json:"title"`
	Message   string           `json:"message"`
	Timestamp time.Time        `json:"timestamp"`
	IsRead    bool             `json:"is_read"`
	PlantID   *uuid.UUID       `json:"plant_id,omitempty"`
	Plant     *Plant           `json:"plant,omitempty"`
}