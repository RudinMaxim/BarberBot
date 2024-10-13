package bot

import (
	"time"

	"github.com/RudinMaxim/BarberBot.git/common"
)

type Service struct {
	repo *Repository
}

func NewClientService(repo *Repository) *Service {
	return &Service{repo: repo}
}

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

// давай реализуем регистрацию. при вызове старт, начинается процес регистрации, нового пользователя. запрашивается контакит. Проверяется, если такой пользователь существует, то ничего не делать так же вызывать welcome_message.

func (s *Service) GetClientByTelegramID(telegramID int64) (*common.Client, error) {
	return s.repo.GetClientByTelegramID(telegramID)
}

func (s *Service) GetClientByTelegram(telegram string) (*common.Client, error) {
	return s.repo.GetClientByTelegram(telegram)
}

func (s *Service) GetClientByPhone(phone string) (*common.Client, error) {
	return s.repo.GetClientByPhone(phone)
}

func (s *Service) GetClientByEmail(email string) (*common.Client, error) {
	return s.repo.GetClientByEmail(email)
}
