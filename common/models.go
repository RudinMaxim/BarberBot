package common

import (
	"time"

	"github.com/google/uuid"
)

// Client модель клиента
type Client struct {
	UUID         uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"uuid"`
	Name         string    `gorm:"not null" json:"name"`
	Phone        string    `gorm:"uniqueIndex" json:"phone"`
	Telegram     string    `json:"telegram"`
	TelegramID   int64     `json:"telegram_id"`
	Email        string    `json:"email"`
	RegisteredAt time.Time `json:"registered_at"`
	LastVisit    time.Time `json:"last_visit"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
	IsActive     bool      `gorm:"default:true" json:"is_active"`
}

// Appointment модель записи на услугу
type Appointment struct {
	ID        int64
	ClientID  int64
	Date      time.Time
	Service   string
	Status    string // "scheduled", "completed", "cancelled"
	CreatedAt time.Time
	UpdatedAt time.Time
}
