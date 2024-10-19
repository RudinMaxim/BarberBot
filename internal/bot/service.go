package bot

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/RudinMaxim/BarberBot.git/common"
	"github.com/google/uuid"
)

type Service struct {
	repo        *Repository
	bookingData map[int64]*BookingData
	mutex       sync.RWMutex
}

type BookingData struct {
	ServiceID uuid.UUID
	Date      time.Time
}

func NewService(repo *Repository) *Service {
	return &Service{
		repo:        repo,
		bookingData: make(map[int64]*BookingData),
	}
}

// ================Client==================

func (s *Service) CreateClient(client *common.Client) (*common.Client, error) {
	client.RegisteredAt = time.Now()
	client.CreatedAt = time.Now()
	client.UpdatedAt = time.Now()
	client.IsActive = true

	createdClient, err := s.repo.CreateClient(client)
	if err != nil {
		return nil, err
	}

	return createdClient, nil
}

func (s *Service) GetClientByTelegramID(telegramID int64) (*common.Client, error) {
	return s.repo.GetClientByTelegramID(telegramID)
}

func (s *Service) GetClientByPhone(phone string) (*common.Client, error) {
	return s.repo.GetClientByPhone(phone)
}

func (s *Service) GetClientByEmail(email string) (*common.Client, error) {
	return s.repo.GetClientByEmail(email)
}

func (s *Service) GetClientAppointments(telegramID int64) ([]common.Appointment, error) {
	client, err := s.GetClientByTelegramID(telegramID)
	if err != nil {
		return nil, fmt.Errorf("failed to get client: %w", err)
	}

	appointments, err := s.repo.GetAppointmentsByClientID(client.UUID)
	if err != nil {
		fmt.Errorf("failed to get appointments: %w", err)
	}

	return appointments, nil
}

func (s *Service) GetClientScheduledAppointmentsByID(telegramID int64) ([]common.Appointment, error) {
	client, err := s.GetClientByTelegramID(telegramID)
	if err != nil {
		return nil, fmt.Errorf("failed to get client: %w", err)
	}

	appointments, err := s.repo.GetScheduledAppointmentsByClientID(client.UUID)
	if err != nil {
		return nil, fmt.Errorf("failed to get appointments: %w", err)
	}
	return appointments, nil
}

// ===============Service==================

func (s *Service) GetServiceByID(serviceID uuid.UUID) (common.Service, error) {
	return s.repo.GetServiceByID(serviceID)
}

func (s *Service) GetActiveServices() ([]common.Service, error) {
	return s.repo.GetActiveServices()
}

// ===============Appointment==================

func (s *Service) GetAppointmentByID(appointmentID uuid.UUID) (*common.Appointment, error) {
	return s.repo.GetAppointmentByID(appointmentID)
}

func (s *Service) CancelAppointment(telegramID int64, appointmentID uuid.UUID) error {
	client, err := s.GetClientByTelegramID(telegramID)
	if err != nil {
		return fmt.Errorf("failed to get client: %w", err)
	}

	appointment, err := s.repo.GetAppointmentByID(appointmentID)
	if err != nil {
		return fmt.Errorf("failed to get appointment: %w", err)
	}

	if appointment.ClientID != client.UUID {
		return errors.New("appointment does not belong to this client")
	}

	now := time.Now()
	if appointment.StartTime.Before(now) {
		return errors.New("cannot cancel past appointments")
	}
	if appointment.Status == "cancelled" {
		return errors.New("appointment is already cancelled")
	}

	appointment.Status = "cancelled"
	appointment.CancelledAt = time.Now()
	appointment.UpdatedAt = time.Now()

	err = s.repo.UpdateAppointment(appointment)
	if err != nil {
		return fmt.Errorf("failed to update appointment: %w", err)
	}

	return nil
}

func (s *Service) CreateAppointment(userID int64, timeStr string) (*common.Appointment, error) {
	serviceID, err := s.getSelectedServiceForUser(userID)
	if err != nil {
		return nil, err
	}
	date, err := s.getSelectedDateForUser(userID)
	if err != nil {
		return nil, err
	}

	layout := "2006-01-02 15:04"
	startTime, err := time.Parse(layout, fmt.Sprintf("%s %s", date.Format("2006-01-02"), timeStr))
	if err != nil {
		return nil, err
	}

	service, err := s.GetServiceByID(serviceID)
	if err != nil {
		return nil, err
	}

	endTime := startTime.Add(time.Duration(service.Duration) * time.Minute)

	client, err := s.GetClientByTelegramID(userID)
	if err != nil {
		return nil, err
	}

	appointment := &common.Appointment{
		ClientID:   client.UUID,
		StartTime:  startTime,
		EndTime:    endTime,
		Name:       service.Name,
		TotalPrice: service.Price,
		Status:     "scheduled",
		Services:   []common.Service{service},
	}

	err = s.repo.CreateAppointment(appointment)
	if err != nil {
		return nil, err
	}

	return appointment, nil
}

// ===============WorkingHours==================

func (s *Service) GetWorkingHoursAvailableDates() ([]time.Time, error) {
	workingHours, err := s.repo.GetWorkingHoursAvailableDates()

	if err != nil {
		fmt.Println("error:", err)
	}

	now := time.Now()

	var availableDates []time.Time
	for i := 0; i < POSSIBLE_RECORDS; i++ {
		date := now.AddDate(0, 0, i)
		dayOfWeek := int(date.Weekday())

		for _, wh := range workingHours {
			if wh.DayOfWeek == dayOfWeek {
				availableDates = append(availableDates, time.Date(date.Year(), date.Month(), date.Day(), wh.StartTime.Hour(), wh.StartTime.Minute(), 0, 0, date.Location()))
				break
			}
		}
	}

	return availableDates, nil
}

func (s *Service) GetWorkingHoursAvailableSlots(serviceIDs []uuid.UUID, date time.Time) ([]time.Time, error) {
	workingHours, err := s.repo.GetWorkingHours(int(date.Weekday()))
	if err != nil {
		return nil, fmt.Errorf("failed to get working hours: %w", err)
	}

	appointments, err := s.repo.GetAppointmentsForDate(date)
	if err != nil {
		return nil, fmt.Errorf("failed to get appointments: %w", err)
	}

	services, err := s.repo.GetServicesByIDs(serviceIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to get services: %w", err)
	}

	totalDuration := 0
	for _, service := range services {
		totalDuration += service.Duration
	}

	availableSlots := []time.Time{}
	currentTime := time.Date(date.Year(), date.Month(), date.Day(), workingHours.StartTime.Hour(), workingHours.StartTime.Minute(), 0, 0, date.Location())
	endTime := time.Date(date.Year(), date.Month(), date.Day(), workingHours.EndTime.Hour(), workingHours.EndTime.Minute(), 0, 0, date.Location())

	for currentTime.Add(time.Duration(totalDuration)*time.Minute).Before(endTime) || currentTime.Add(time.Duration(totalDuration)*time.Minute).Equal(endTime) {
		isAvailable := true
		potentialEndTime := currentTime.Add(time.Duration(totalDuration) * time.Minute)

		for _, appointment := range appointments {
			if (currentTime.Before(appointment.EndTime) && potentialEndTime.After(appointment.StartTime)) ||
				(currentTime.Equal(appointment.StartTime) || potentialEndTime.Equal(appointment.EndTime)) {
				isAvailable = false
				break
			}
		}

		if isAvailable {
			availableSlots = append(availableSlots, currentTime)
		}

		currentTime = currentTime.Add(30 * time.Minute)
	}

	return availableSlots, nil
}

// =================================

func (s *Service) SaveSelectedService(userID int64, serviceID string) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	uuid, err := uuid.Parse(serviceID)
	if err != nil {
		return fmt.Errorf("invalid service ID: %v", err)
	}

	if _, ok := s.bookingData[userID]; !ok {
		s.bookingData[userID] = &BookingData{}
	}
	s.bookingData[userID].ServiceID = uuid
	return nil
}

func (s *Service) SaveSelectedDate(userID int64, date time.Time) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if _, ok := s.bookingData[userID]; !ok {
		s.bookingData[userID] = &BookingData{}
	}
	s.bookingData[userID].Date = date
	return nil
}

func (s *Service) getSelectedServiceForUser(userID int64) (uuid.UUID, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	if data, ok := s.bookingData[userID]; ok && data.ServiceID != uuid.Nil {
		return data.ServiceID, nil
	}
	return uuid.Nil, fmt.Errorf("no service selected for user %d", userID)
}

func (s *Service) getSelectedDateForUser(userID int64) (time.Time, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	if data, ok := s.bookingData[userID]; ok && !data.Date.IsZero() {
		return data.Date, nil
	}
	return time.Time{}, fmt.Errorf("no date selected for user %d", userID)
}

func (s *Service) ClearBookingData(userID int64) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	delete(s.bookingData, userID)
}
