package appointments

import (
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

func (r *Repository) GetAvailableSlots(serviceIDs []uuid.UUID, date time.Time, bufferMinutes int) ([]time.Time, error) {
	// Получаем рабочие часы для данного дня недели
	var workingHours common.WorkingHours
	if err := r.db.Where("day_of_week = ?", int(date.Weekday())).First(&workingHours).Error; err != nil {
		return nil, err
	}

	// Получаем все записи на выбранную дату
	var appointments []common.Appointment
	if err := r.db.Where("start_time BETWEEN ? AND ?", date, date.Add(24*time.Hour)).Find(&appointments).Error; err != nil {
		return nil, err
	}

	// Рассчитываем общую длительность выбранных услуг
	var totalDuration time.Duration
	for _, serviceID := range serviceIDs {
		var service common.Service
		if err := r.db.First(&service, serviceID).Error; err != nil {
			return nil, err
		}
		totalDuration += time.Duration(service.Duration) * time.Minute
	}

	// Генерируем доступные слоты
	availableSlots := []time.Time{}
	currentTime := time.Date(date.Year(), date.Month(), date.Day(), workingHours.StartTime.Hour(), workingHours.StartTime.Minute(), 0, 0, date.Location())
	endTime := time.Date(date.Year(), date.Month(), date.Day(), workingHours.EndTime.Hour(), workingHours.EndTime.Minute(), 0, 0, date.Location())

	for currentTime.Add(totalDuration).Before(endTime) || currentTime.Add(totalDuration).Equal(endTime) {
		isAvailable := true
		for _, appointment := range appointments {
			if (currentTime.Before(appointment.EndTime) && currentTime.Add(totalDuration).After(appointment.StartTime)) ||
				(currentTime.Equal(appointment.StartTime) || currentTime.Add(totalDuration).Equal(appointment.EndTime)) {
				isAvailable = false
				break
			}
		}

		if isAvailable {
			availableSlots = append(availableSlots, currentTime)
		}

		currentTime = currentTime.Add(15 * time.Minute)
	}

	return availableSlots, nil
}

func (r *Repository) CreateAppointment(appointment *common.Appointment) error {
	return r.db.Create(appointment).Error
}

func (r *Repository) GetClientAppointments(clientID uuid.UUID) ([]common.Appointment, error) {
	var appointments []common.Appointment
	err := r.db.Where("client_id = ?", clientID).Find(&appointments).Error
	return appointments, err
}

func (r *Repository) GetAppointmentByID(appointmentID uuid.UUID) (*common.Appointment, error) {
	var appointment common.Appointment
	err := r.db.First(&appointment, appointmentID).Error
	return &appointment, err
}

func (r *Repository) UpdateAppointment(appointment *common.Appointment) error {
	return r.db.Save(appointment).Error
}

func (r *Repository) GetServiceByID(serviceID uuid.UUID) (common.Service, error) {
	var service common.Service
	err := r.db.First(&service, serviceID).Error
	return service, err
}

func (r *Repository) GetServiceList() ([]common.Service, error) {
	var services []common.Service
	err := r.db.Find(&services).Error
	return services, err
}
