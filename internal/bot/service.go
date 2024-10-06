package bot

import (
	"fmt"
	"strings"
	"time"

	"github.com/RudinMaxim/BarberBot.git/common"
	"github.com/RudinMaxim/BarberBot.git/config"
	"github.com/RudinMaxim/BarberBot.git/helper"
)

type RegistrationState int

const (
	StateNone RegistrationState = iota
	StateAwaitingName
	StateAwaitingPhone
)

type RegistrationData struct {
	State RegistrationState
	Name  string
	Phone string
}

type RegistrationResult struct {
	Message string
	Done    bool
}

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{
		repo: repo,
	}
}

func (s *Service) HandleRegistration(userID int64, input string) RegistrationResult {
	state := s.repo.GetRegistrationState(userID)

	switch state.State {
	case StateNone:
		s.repo.SetRegistrationState(userID, &RegistrationData{State: StateAwaitingName})
		return RegistrationResult{Message: helper.GetText("registration_start"), Done: false}

	case StateAwaitingName:
		if !helper.IsValidName(input) {
			return RegistrationResult{Message: "Пожалуйста, введите корректное имя (только буквы, минимум 2 символа).", Done: false}
		}
		state.Name = input
		state.State = StateAwaitingPhone
		s.repo.SetRegistrationState(userID, state)
		msg := helper.GetText("registration_name_received")
		msg = strings.Replace(msg, "{{.Name}}", input, -1)
		return RegistrationResult{Message: msg, Done: false}

	case StateAwaitingPhone:
		normalizedPhone := helper.NormalizePhoneNumber(input)
		if !helper.IsValidPhone(normalizedPhone) {
			return RegistrationResult{Message: "Пожалуйста, введите корректный номер телефона.", Done: false}
		}
		state.Phone = normalizedPhone
		return s.completeRegistration(userID, state)

	default:
		return RegistrationResult{Message: helper.GetText("registration_error"), Done: false}
	}
}

func (s *Service) completeRegistration(userID int64, state *RegistrationData) RegistrationResult {
	now := time.Now()
	client := common.Client{
		ID:           userID,
		Name:         state.Name,
		Phone:        state.Phone,
		RegisteredAt: now,
		CreatedAt:    now,
		UpdatedAt:    now,
	}
	err := s.repo.SaveClient(client)
	if err != nil {
		config.LogAction(fmt.Sprintf("Ошибка при сохранении клиента: %v", err))
		return RegistrationResult{Message: helper.GetText("registration_error"), Done: false}
	}
	s.repo.ClearRegistrationState(userID)
	config.LogAction(fmt.Sprintf("Клиент успешно зарегистрирован: ID=%d, Имя=%s, Телефон=%s", userID, state.Name, state.Phone))

	msg := helper.GetText("registration_complete")
	msg = strings.Replace(msg, "{{.Name}}", state.Name, -1)
	return RegistrationResult{Message: msg, Done: true}
}

func (s *Service) GetRegistrationState(userID int64) *RegistrationData {
	return s.repo.GetRegistrationState(userID)
}
