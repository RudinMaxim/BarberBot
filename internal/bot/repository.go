package bot

import (
	"fmt"
	"time"

	"github.com/RudinMaxim/BarberBot.git/common"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

// ===============Client===================

func (r *Repository) CreateClient(client *common.Client) (*common.Client, error) {
	err := r.db.Create(client).Error
	return client, err
}

func (r *Repository) GetClientBy(field string, value interface{}) (*common.Client, error) {
	var client common.Client
	err := r.db.Where(fmt.Sprintf("%s = ?", field), value).First(&client).Error
	return &client, err
}

// ===============Appointment===================

func (r *Repository) CreateAppointment(appointment *common.Appointment) error {
	return r.db.Create(appointment).Error
}

func (r *Repository) GetAppointmentByID(appointmentID uuid.UUID) (*common.Appointment, error) {
	var appointment common.Appointment
	err := r.db.Preload("Services").First(&appointment, appointmentID).Error
	return &appointment, err
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
	var service common.Service
	err := r.db.First(&service, serviceID).Error
	return service, err
}

func (r *Repository) GetActiveServices() ([]common.Service, error) {
	var services []common.Service
	err := r.db.Where("is_active = ?", true).Find(&services).Error
	return services, err
}

func (r *Repository) GetServicesByIDs(serviceIDs []uuid.UUID) ([]common.Service, error) {
	var services []common.Service
	err := r.db.Where("uuid IN ?", serviceIDs).Find(&services).Error
	return services, err
}

// ===============WorkingHours===================

func (r *Repository) GetWorkingHoursByDayOfWeek(dayOfWeek int) (*common.WorkingHours, error) {
	var workingHours common.WorkingHours
	err := r.db.Where("day_of_week = ? AND is_active = ?", dayOfWeek, true).First(&workingHours).Error
	return &workingHours, err
}

func (r *Repository) GetWorkingHoursAvailableDates() ([]common.WorkingHours, error) {
	var workingHours []common.WorkingHours
	err := r.db.Where("is_active = ?", true).Find(&workingHours).Error
	return workingHours, err
}

func (r *Repository) GetWorkingHours() ([]common.WorkingHours, error) {
	var workingHours common.WorkingHours
	err := r.db.Find(&workingHours).Error
	return []common.WorkingHours{workingHours}, err
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
