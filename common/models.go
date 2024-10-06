package common

import "time"

// Client модель клиента
type Client struct {
	ID           int64
	Name         string
	Phone        string
	Telegram     string
	Email        string
	RegisteredAt time.Time
	LastVisit    time.Time
	CreatedAt    time.Time
	UpdatedAt    time.Time
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
