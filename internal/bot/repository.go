package bot

import (
	"context"
	"fmt"
	"time"

	"github.com/RudinMaxim/BarberBot.git/common"
	"github.com/RudinMaxim/BarberBot.git/database"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Repository struct {
	db    *gorm.DB
	cache *database.RedisCache
}

func NewRepository(db *gorm.DB, cache *database.RedisCache) *Repository {
	return &Repository{
		db:    db,
		cache: cache,
	}
}

// ===============Client===================

func (r *Repository) CreateClient(client *common.Client) (*common.Client, error) {
	err := r.db.Create(client).Error
	return client, err
}

func (r *Repository) GetClientBy(field string, value interface{}) (*common.Client, error) {
	ctx := context.Background()
	cacheKey := fmt.Sprintf("client:%s:%v", field, value)

	var client common.Client

	// Попытка получить данные из кэша
	err := r.cache.Get(ctx, cacheKey, &client)
	if err == nil {
		return &client, nil
	}

	// Запрос к базе данных, если данных нет в кэше
	err = r.db.Where(fmt.Sprintf("%s = ?", field), value).First(&client).Error
	if err != nil {
		return nil, err
	}

	// Сохраняем результат в кэш на 1 час
	cacheDuration := time.Hour
	if err = r.cache.Set(ctx, cacheKey, client, cacheDuration); err != nil {
		return nil, fmt.Errorf("failed to cache data: %w", err)
	}

	return &client, nil
}

// ===============Appointment===================

func (r *Repository) CreateAppointment(appointment *common.Appointment) error {
	return r.db.Create(appointment).Error
}

func (r *Repository) GetAppointmentByID(appointmentID uuid.UUID) (*common.Appointment, error) {
	var appointment common.Appointment
	err := r.db.Preload("Services").First(&appointment, appointmentID).Error
	if err != nil {
		return nil, err
	}
	return &appointment, nil
}

func (r *Repository) GetAppointmentsByClientID(clientID uuid.UUID) ([]common.Appointment, error) {
	var appointments []common.Appointment
	err := r.db.Where("client_id = ?", clientID).Find(&appointments).Error
	return appointments, err
}

func (r *Repository) UpdateAppointment(appointment *common.Appointment) error {
	return r.db.Save(appointment).Error
}

func (r *Repository) GetAppointmentsForDate(date time.Time) ([]common.Appointment, error) {
	var appointments []common.Appointment
	err := r.db.Preload("Services").
		Where("start_time BETWEEN ? AND ?", date, date.Add(24*time.Hour)).
		Find(&appointments).Error
	return appointments, err
}

func (r *Repository) GetScheduledAppointmentsByClientID(clientID uuid.UUID) ([]common.Appointment, error) {
	var appointments []common.Appointment
	err := r.db.Where("client_id = ? AND status = ?", clientID, "scheduled").Find(&appointments).Error
	return appointments, err
}

// ===============Service===================

func (r *Repository) GetServiceByID(serviceID uuid.UUID) (common.Service, error) {
	ctx := context.Background()
	cacheKey := fmt.Sprintf("service:%s", serviceID)

	var service common.Service

	// Попытка получить данные из кэша
	err := r.cache.Get(ctx, cacheKey, &service)
	if err == nil {
		return service, nil
	}

	// Запрос к базе данных, если данных нет в кэше
	err = r.db.First(&service, serviceID).Error
	if err != nil {
		return common.Service{}, err
	}

	// Сохраняем результат в кэш на 1 час
	cacheDuration := time.Hour
	if err = r.cache.Set(ctx, cacheKey, service, cacheDuration); err != nil {
		return common.Service{}, fmt.Errorf("failed to cache data: %w", err)
	}

	return service, nil
}

func (r *Repository) GetActiveServices() ([]common.Service, error) {
	ctx := context.Background()
	cacheKey := "active_services"

	var services []common.Service

	// Попытка получить данные из кэша
	err := r.cache.Get(ctx, cacheKey, &services)
	if err == nil {
		return services, nil
	}

	// Запрос к базе данных, если данных нет в кэше
	err = r.db.Where("is_active = ?", true).Find(&services).Error
	if err != nil {
		return nil, err
	}

	// Сохраняем результат в кэш на 1 час
	cacheDuration := time.Hour
	if err = r.cache.Set(ctx, cacheKey, services, cacheDuration); err != nil {
		return nil, fmt.Errorf("failed to cache data: %w", err)
	}

	return services, nil
}

func (r *Repository) GetServicesByIDs(serviceIDs []uuid.UUID) ([]common.Service, error) {
	ctx := context.Background()
	cacheKey := fmt.Sprintf("services:ids:%v", serviceIDs)

	var services []common.Service

	// Попытка получить данные из кэша
	err := r.cache.Get(ctx, cacheKey, &services)
	if err == nil {
		return services, nil
	}

	// Запрос к базе данных, если данных нет в кэше
	err = r.db.Where("uuid IN ?", serviceIDs).Find(&services).Error
	if err != nil {
		return nil, err
	}

	// Сохраняем результат в кэш на 1 час
	cacheDuration := time.Hour
	if err = r.cache.Set(ctx, cacheKey, services, cacheDuration); err != nil {
		return nil, fmt.Errorf("failed to cache data: %w", err)
	}

	return services, nil
}

// ===============WorkingHours===================

func (r *Repository) GetWorkingHoursByDayOfWeek(dayOfWeek int) (*common.WorkingHours, error) {
	var workingHours common.WorkingHours
	err := r.db.Where("day_of_week = ? AND is_active = ?", dayOfWeek, true).First(&workingHours).Error
	return &workingHours, err
}

func (r *Repository) GetWorkingHoursAvailableDates() ([]common.WorkingHours, error) {
	ctx := context.Background()
	cacheKey := "working_hours:available"

	var workingHours []common.WorkingHours

	err := r.cache.Get(ctx, cacheKey, &workingHours)
	if err == nil {
		return workingHours, nil
	}

	err = r.db.Where("is_active = ?", true).Find(&workingHours).Error
	if err != nil {
		return nil, err
	}

	cacheDuration := 24 * time.Hour
	if err = r.cache.Set(ctx, cacheKey, workingHours, cacheDuration); err != nil {
		return nil, fmt.Errorf("failed to cache working hours: %w", err)
	}

	return workingHours, nil
}

func (r *Repository) GetWorkingHours() ([]common.WorkingHours, error) {
	ctx := context.Background()
	cacheKey := "working_hours:all"

	var workingHours common.WorkingHours

	err := r.cache.Get(ctx, cacheKey, &workingHours)
	if err == nil {
		return []common.WorkingHours{workingHours}, nil
	}

	err = r.db.Find(&workingHours).Error
	if err != nil {
		return nil, err
	}

	cacheDuration := 24 * time.Hour
	if err = r.cache.Set(ctx, cacheKey, workingHours, cacheDuration); err != nil {
		return nil, fmt.Errorf("failed to cache working hours: %w", err)
	}

	return []common.WorkingHours{workingHours}, nil
}

func (r *Repository) SaveCalendarEventID(appointmentID uuid.UUID, eventID string) error {
	return r.db.Model(&common.Appointment{}).
		Where("uuid = ?", appointmentID).
		Update("calendar_event_id", eventID).Error
}

func (r *Repository) GetCalendarEventID(appointmentID uuid.UUID) (string, error) {
	var appointment common.Appointment
	result := r.db.Select("calendar_event_id").
		Where("uuid = ?", appointmentID).
		First(&appointment)

	if result.Error != nil {
		return "", result.Error
	}

	return appointment.CalendarEventID, nil
}
