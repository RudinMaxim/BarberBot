package sheets

// import (
// 	"context"
// 	"fmt"
// 	"strconv"
// 	"time"

// 	"github.com/RudinMaxim/BarberBot.git/common"

// 	"google.golang.org/api/sheets/v4"
// )

// type Repository struct {
// 	srv           *sheets.Service
// 	spreadsheetID string
// 	userSheet     string
// 	ctx           context.Context
// }

// func NewRepository(ctx context.Context) *Repository {
// 	return &Repository{
// 		userSheet: "Users",
// 		ctx:       ctx,
// 	}
// }

// func (r *Repository) Init(srv *sheets.Service, spreadsheetID string) error {
// 	r.srv = srv
// 	r.spreadsheetID = spreadsheetID

// 	// Verify the spreadsheet and sheet exist
// 	_, err := r.srv.Spreadsheets.Get(spreadsheetID).Do()
// 	if err != nil {
// 		return fmt.Errorf("failed to access spreadsheet: %v", err)
// 	}

// 	// Check if the Users sheet exists, create if not
// 	sheet, err := r.srv.Spreadsheets.Get(spreadsheetID).Fields("sheets(properties(title))").Do()
// 	if err != nil {
// 		return fmt.Errorf("failed to get sheet info: %v", err)
// 	}

// 	userSheetExists := false
// 	for _, s := range sheet.Sheets {
// 		if s.Properties.Title == r.userSheet {
// 			userSheetExists = true
// 			break
// 		}
// 	}

// 	if !userSheetExists {
// 		// Create the Users sheet
// 		addSheetRequest := &sheets.Request{
// 			AddSheet: &sheets.AddSheetRequest{
// 				Properties: &sheets.SheetProperties{
// 					Title: r.userSheet,
// 				},
// 			},
// 		}

// 		batchUpdateRequest := &sheets.BatchUpdateSpreadsheetRequest{
// 			Requests: []*sheets.Request{addSheetRequest},
// 		}

// 		_, err := r.srv.Spreadsheets.BatchUpdate(spreadsheetID, batchUpdateRequest).Context(r.ctx).Do()
// 		if err != nil {
// 			return fmt.Errorf("failed to create Users sheet: %v", err)
// 		}

// 		// Add headers to the new sheet
// 		headers := []interface{}{"ID", "Name", "Phone", "Telegram", "Email", "RegisteredAt", "LastVisit", "CreatedAt", "UpdatedAt"}
// 		valueRange := &sheets.ValueRange{
// 			Values: [][]interface{}{headers},
// 		}
// 		_, err = r.srv.Spreadsheets.Values.Append(spreadsheetID, r.userSheet+"!A1", valueRange).ValueInputOption("USER_ENTERED").Context(r.ctx).Do()
// 		if err != nil {
// 			return fmt.Errorf("failed to add headers to Users sheet: %v", err)
// 		}
// 	}

// 	return nil
// }

// func (r *Repository) CreateUser(client common.Client) error {
// 	values := []interface{}{
// 		client.ID,
// 		client.Name,
// 		client.Phone,
// 		client.Telegram,
// 		client.Email,
// 		client.RegisteredAt.Format(time.RFC3339),
// 		client.LastVisit.Format(time.RFC3339),
// 		time.Now().Format(time.RFC3339),
// 		time.Now().Format(time.RFC3339),
// 	}
// 	valueRange := &sheets.ValueRange{
// 		Values: [][]interface{}{values},
// 	}
// 	_, err := r.srv.Spreadsheets.Values.Append(r.spreadsheetID, r.userSheet, valueRange).ValueInputOption("USER_ENTERED").Context(r.ctx).Do()
// 	return err
// }

// func (r *Repository) ReadUser(id int64) (common.Client, error) {
// 	readRange := r.userSheet + "!A:I"
// 	resp, err := r.srv.Spreadsheets.Values.Get(r.spreadsheetID, readRange).Context(r.ctx).Do()
// 	if err != nil {
// 		return common.Client{}, err
// 	}
// 	for _, row := range resp.Values[1:] { // Skip header row
// 		if rowID, _ := strconv.ParseInt(row[0].(string), 10, 64); rowID == id {
// 			registeredAt, _ := time.Parse(time.RFC3339, row[5].(string))
// 			lastVisit, _ := time.Parse(time.RFC3339, row[6].(string))
// 			createdAt, _ := time.Parse(time.RFC3339, row[7].(string))
// 			updatedAt, _ := time.Parse(time.RFC3339, row[8].(string))
// 			return common.Client{
// 				ID:           id,
// 				Name:         row[1].(string),
// 				Phone:        row[2].(string),
// 				Telegram:     row[3].(string),
// 				Email:        row[4].(string),
// 				RegisteredAt: registeredAt,
// 				LastVisit:    lastVisit,
// 				CreatedAt:    createdAt,
// 				UpdatedAt:    updatedAt,
// 			}, nil
// 		}
// 	}
// 	return common.Client{}, fmt.Errorf("user not found")
// }

// func (r *Repository) UpdateUser(client common.Client) error {
// 	readRange := r.userSheet + "!A:I"
// 	resp, err := r.srv.Spreadsheets.Values.Get(r.spreadsheetID, readRange).Context(r.ctx).Do()
// 	if err != nil {
// 		return err
// 	}
// 	for i, row := range resp.Values[1:] { // Skip header row
// 		if rowID, _ := strconv.ParseInt(row[0].(string), 10, 64); rowID == client.ID {
// 			updateRange := fmt.Sprintf("%s!A%d:I%d", r.userSheet, i+2, i+2)
// 			values := []interface{}{
// 				client.ID,
// 				client.Name,
// 				client.Phone,
// 				client.Telegram,
// 				client.Email,
// 				client.RegisteredAt.Format(time.RFC3339),
// 				client.LastVisit.Format(time.RFC3339),
// 				row[7], // Keep original CreatedAt
// 				time.Now().Format(time.RFC3339),
// 			}
// 			valueRange := &sheets.ValueRange{
// 				Values: [][]interface{}{values},
// 			}
// 			_, err := r.srv.Spreadsheets.Values.Update(r.spreadsheetID, updateRange, valueRange).ValueInputOption("USER_ENTERED").Context(r.ctx).Do()
// 			return err
// 		}
// 	}
// 	return fmt.Errorf("user not found")
// }

// func (r *Repository) DeleteUser(id int64) error {
// 	readRange := r.userSheet + "!A:I"
// 	resp, err := r.srv.Spreadsheets.Values.Get(r.spreadsheetID, readRange).Context(r.ctx).Do()
// 	if err != nil {
// 		return err
// 	}
// 	for i, row := range resp.Values[1:] { // Skip header row
// 		if rowID, _ := strconv.ParseInt(row[0].(string), 10, 64); rowID == id {
// 			deleteRange := fmt.Sprintf("%s!A%d:I%d", r.userSheet, i+2, i+2)
// 			_, err := r.srv.Spreadsheets.Values.Clear(r.spreadsheetID, deleteRange, &sheets.ClearValuesRequest{}).Context(r.ctx).Do()
// 			return err
// 		}
// 	}
// 	return fmt.Errorf("user not found")
// }
