package appointments

import (
	"errors"
	"time"

	"github.com/RudinMaxim/BarberBot.git/common"
	"github.com/google/uuid"
)

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) GetServiceList() ([]common.Service, error) {
	return s.repo.GetServiceList()
}

func (s *Service) GetServiceByID(serviceID uuid.UUID) (common.Service, error) {
	return s.repo.GetServiceByID(serviceID)
}

func (s *Service) GetAvailableSlots(serviceIDs []uuid.UUID, date time.Time, bufferMinutes int) ([]time.Time, error) {
	return s.repo.GetAvailableSlots(serviceIDs, date, bufferMinutes)
}

func (s *Service) CreateAppointment(clientID uuid.UUID, serviceIDs []uuid.UUID, startTime time.Time) (*common.Appointment, error) {
	// Calculate total duration and price
	var totalDuration int
	var totalPrice float64
	for _, serviceID := range serviceIDs {
		service, err := s.repo.GetServiceByID(serviceID)
		if err != nil {
			return nil, err
		}
		totalDuration += service.Duration
		totalPrice += service.Price
	}

	endTime := startTime.Add(time.Duration(totalDuration) * time.Minute)

	appointment := &common.Appointment{
		UUID:       uuid.New(),
		ClientID:   clientID,
		ServiceIDs: serviceIDs,
		StartTime:  startTime,
		EndTime:    endTime,
		TotalPrice: totalPrice,
		Status:     "scheduled",
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	err := s.repo.CreateAppointment(appointment)
	if err != nil {
		return nil, err
	}

	return appointment, nil
}

func (s *Service) GetClientAppointments(clientID uuid.UUID) ([]common.Appointment, error) {
	return s.repo.GetClientAppointments(clientID)
}

func (s *Service) CancelAppointment(appointmentID uuid.UUID) error {
	appointment, err := s.repo.GetAppointmentByID(appointmentID)
	if err != nil {
		return err
	}

	if appointment.Status != "scheduled" {
		return errors.New("appointment cannot be cancelled")
	}

	appointment.Status = "cancelled"
	appointment.CancelledAt = time.Now()
	appointment.UpdatedAt = time.Now()

	return s.repo.UpdateAppointment(appointment)
}

func (s *Service) RescheduleAppointment(appointmentID uuid.UUID, newStartTime time.Time) error {
	appointment, err := s.repo.GetAppointmentByID(appointmentID)
	if err != nil {
		return err
	}

	if appointment.Status != "scheduled" {
		return errors.New("appointment cannot be rescheduled")
	}

	// Recalculate end time
	duration := appointment.EndTime.Sub(appointment.StartTime)
	newEndTime := newStartTime.Add(duration)

	appointment.StartTime = newStartTime
	appointment.EndTime = newEndTime
	appointment.UpdatedAt = time.Now()

	return s.repo.UpdateAppointment(appointment)
}
