package bot

import (
	"fmt"

	"github.com/RudinMaxim/BarberBot.git/common"
	"gorm.io/gorm"
)

type Repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) CreateClient(client *common.Client) error {
	return r.db.Create(client).Error
}

func (r *Repository) GetClientBy(field string, value interface{}) (*common.Client, error) {
	var client common.Client
	err := r.db.Where(field+" = ?", value).First(&client).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get category by %w: %w", field, err)
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
