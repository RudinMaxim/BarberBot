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

// Service модель услуги
type Service struct {
	UUID      uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"uuid"`
	Name      string    `gorm:"type:varchar(255);not null" json:"name"`   // Название услуги (например, "Мужская стрижка")
	Duration  int       `gorm:"not null" json:"duration"`                 // Длительность в минутах
	Price     float64   `gorm:"type:decimal(10,2);not null" json:"price"` // Цена услуги
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	IsActive  bool      `gorm:"default:true" json:"is_active"`
}

// WorkingHours модель рабочих часов
type WorkingHours struct {
	UUID      uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"uuid"`
	DayOfWeek int       `json:"day_of_week"`
	StartTime time.Time `gorm:"type:timestamp;not null" json:"start_time"`
	EndTime   time.Time `gorm:"type:timestamp;not null" json:"end_time"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	IsActive  bool      `json:"is_active"`
}

type Appointment struct {
	UUID        uuid.UUID   `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"uuid"`
	ClientID    uuid.UUID   `gorm:"type:uuid;not null" json:"client_id"`
	ServiceIDs  []uuid.UUID `gorm:"type:uuid[]" json:"service_ids"`
	StartTime   time.Time   `gorm:"type:timestamp" json:"start_time"`
	EndTime     time.Time   `gorm:"type:timestamp" json:"end_time"`
	Name        string      `gorm:"not null" json:"name"`
	TotalPrice  float64     `gorm:"type:decimal(10,2);not null" json:"total_price"`
	Status      string      `gorm:"type:varchar(20);not null" json:"status"` // "scheduled", "completed", "cancelled"
	CreatedAt   time.Time   `json:"created_at"`
	UpdatedAt   time.Time   `json:"updated_at"`
	CancelledAt time.Time   `json:"cancelled_at,omitempty"`
}
