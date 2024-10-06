package sheets

// import (
// 	"context"
// 	"fmt"

// 	"github.com/RudinMaxim/BarberBot.git/common"
// 	"golang.org/x/oauth2/google"
// 	"google.golang.org/api/option"
// 	"google.golang.org/api/sheets/v4"
// )

// type Handler struct {
// 	service       *Service
// 	spreadsheetID string
// }

// func NewHandler(ctx context.Context, service *Service, credentialsFile, spreadsheetID string) (*Handler, error) {
// 	credentials, err := google.CredentialsFromJSON(ctx, []byte(credentialsFile), sheets.SpreadsheetsScope)
// 	if err != nil {
// 		return nil, fmt.Errorf("unable to parse credentials: %v", err)
// 	}

// 	srv, err := sheets.NewService(ctx, option.WithCredentials(credentials))
// 	if err != nil {
// 		return nil, fmt.Errorf("unable to retrieve Sheets client: %v", err)
// 	}

// 	if err := service.repo.Init(srv, spreadsheetID); err != nil {
// 		return nil, fmt.Errorf("failed to initialize repository: %v", err)
// 	}

// 	return &Handler{
// 		service:       service,
// 		spreadsheetID: spreadsheetID,
// 	}, nil
// }

// func (h *Handler) HandleCreateUser(client common.Client) error {
// 	return h.service.CreateUser(client)
// }

// func (h *Handler) HandleReadUser(id int64) (common.Client, error) {
// 	return h.service.ReadUser(id)
// }

// func (h *Handler) HandleUpdateUser(client common.Client) error {
// 	return h.service.UpdateUser(client)
// }

// func (h *Handler) HandleDeleteUser(id int64) error {
// 	return h.service.DeleteUser(id)
// }
