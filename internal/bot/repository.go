package bot

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/RudinMaxim/BarberBot.git/common"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Repository struct {
	db       *gorm.DB
	services []common.Service
}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) CreateClient(client *common.Client) (*common.Client, error) {
	err := r.db.Create(client).Error
	if err != nil {
		return nil, err
	}

	return client, nil
}

func (r *Repository) GetClientBy(field string, value interface{}) (*common.Client, error) {
	var client common.Client
	err := r.db.Session(&gorm.Session{PrepareStmt: true}).Where(field+" = ?", value).First(&client).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get client by %s: %w", field, err)
	}
	return &client, nil
}

func (r *Repository) GetClientByTelegramID(telegramID int64) (*common.Client, error) {
	return r.GetClientBy("telegram_id", telegramID)
}

func (r *Repository) GetClientByTelegram(telegram string) (*common.Client, error) {
	return r.GetClientBy("telegram", telegram)
}

func (r *Repository) GetClientByPhone(phone string) (*common.Client, error) {
	return r.GetClientBy("phone", phone)
}

func (r *Repository) GetClientByEmail(email string) (*common.Client, error) {
	return r.GetClientBy("email", email)
}

func (r *Repository) CreateAppointment(appointment *common.Appointment) error {
	query := `
	INSERT INTO appointments (
		client_id,
		service_ids,
		start_time,
		end_time,
		name,
		total_price,
		status,
		created_at,
		updated_at,
		cancelled_at
	) VALUES (
		$1,
		$2,
		$3,
		$4,
		$5,
		$6,
		$7,
		$8,
		$9,
		$10
	) RETURNING uuid`

	serviceIDsArray := fmt.Sprintf("{%s}", appointment.ServiceIDs[0])
	now := time.Now().Format("2006-01-02 15:04:05-07")

	var uuidStr string
	err := r.db.Raw(query,
		appointment.ClientID,
		serviceIDsArray,
		appointment.StartTime,
		appointment.EndTime,
		appointment.Name,
		appointment.TotalPrice,
		appointment.Status,
		now,
		now,
		sql.NullTime{},
	).Scan(&uuidStr).Error

	if err != nil {
		return fmt.Errorf("failed to execute create appointment query: %w", err)
	}

	appointmentUUID, err := uuid.Parse(uuidStr)
	if err != nil {
		return fmt.Errorf("failed to parse returned UUID: %w", err)
	}

	appointment.UUID = appointmentUUID
	return nil
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

func (r *Repository) GetWorkingHours(dayOfWeek int) (*common.WorkingHours, error) {
	var workingHours common.WorkingHours
	err := r.db.Where("day_of_week = ? AND is_active = ?", dayOfWeek, true).First(&workingHours).Error
	if err != nil {
		fmt.Println("error:", err)
	}
	return &workingHours, nil
}

func (r *Repository) GetAppointmentsForDate(date time.Time) ([]common.Appointment, error) {
	var appointments []common.Appointment
	fmt.Println("date:", date)
	err := r.db.Where("start_time BETWEEN ? AND ?", date, date.Add(24*time.Hour)).Find(&appointments).Error
	if err != nil {
		fmt.Println("error:", err)
	}
	return appointments, nil
}

func (r *Repository) GetServicesByIDs(serviceIDs []uuid.UUID) ([]common.Service, error) {
	var services []common.Service
	err := r.db.Where("uuid IN ?", serviceIDs).Find(&services).Error
	return services, err
}

func (r *Repository) GetActiveServices() ([]common.Service, error) {
	var services []common.Service
	err := r.db.Where("is_active = ?", true).Find(&services).Error

	return services, err
}

func (r *Repository) GetAvailableDates() ([]common.WorkingHours, error) {
	var workingHours []common.WorkingHours
	err := r.db.Where("is_active = ?", true).Find(&workingHours).Error

	return workingHours, err
}
