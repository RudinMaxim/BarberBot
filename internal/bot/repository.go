package bot

import (
	"fmt"

	"github.com/RudinMaxim/BarberBot.git/common"
	"github.com/RudinMaxim/BarberBot.git/config"
)

type Repository struct {
	clients            map[int64]common.Client
	registrationStates map[int64]*RegistrationData
}

func NewRepository() *Repository {
	return &Repository{
		clients:            make(map[int64]common.Client),
		registrationStates: make(map[int64]*RegistrationData),
	}
}

func (r *Repository) SaveClient(client common.Client) error {
	r.clients[client.ID] = client
	config.LogAction(fmt.Sprintf("Сохранен новый клиент: ID=%d, Имя=%s, Телефон=%s", client.ID, client.Name, client.Phone))
	return nil
}

func (r *Repository) GetRegistrationState(userID int64) *RegistrationData {
	state, exists := r.registrationStates[userID]
	if !exists {
		state = &RegistrationData{State: StateNone}
		r.registrationStates[userID] = state
	}
	return state
}

func (r *Repository) SetRegistrationState(userID int64, state *RegistrationData) {
	r.registrationStates[userID] = state
}

func (r *Repository) ClearRegistrationState(userID int64) {
	delete(r.registrationStates, userID)
}
