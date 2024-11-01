package calendar

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/RudinMaxim/BarberBot.git/common"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/calendar/v3"
	"google.golang.org/api/option"
)

const (
	defaultTimeZone     = "Asia/Yekaterinburg"
	alternativeTimeZone = "+05:00"
)

type GoogleCalendarService struct {
	client     *calendar.Service
	calendarID string
}

func getClient(config *oauth2.Config) *http.Client {
	tokFile := "credentials/token.json"
	tok, err := tokenFromFile(tokFile)
	if err != nil {
		tok = getTokenFromWeb(config)
		saveToken(tokFile, tok)
	}
	return config.Client(context.Background(), tok)
}

func getTokenFromWeb(config *oauth2.Config) *oauth2.Token {
	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	fmt.Printf("Перейдите по ссылке для авторизации:\n%v\n", authURL)

	var authCode string
	if _, err := fmt.Scan(&authCode); err != nil {
		log.Fatalf("Ошибка чтения кода авторизации: %v", err)
	}

	tok, err := config.Exchange(context.TODO(), authCode)
	if err != nil {
		log.Fatalf("Не удалось получить токен: %v", err)
	}
	return tok
}

func tokenFromFile(file string) (*oauth2.Token, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	tok := &oauth2.Token{}
	err = json.NewDecoder(f).Decode(tok)
	return tok, err
}

func saveToken(path string, token *oauth2.Token) {
	fmt.Printf("Сохранение токена в: %s\n", path)
	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		log.Fatalf("Не удалось сохранить токен: %v", err)
	}
	defer f.Close()
	json.NewEncoder(f).Encode(token)
}

func NewGoogleCalendarService() (*GoogleCalendarService, error) {
	ctx := context.Background()
	b, err := os.ReadFile("credentials/credentials.json")

	if err != nil {
		return nil, fmt.Errorf("не удалось прочитать файл credentials: %w", err)
	}

	config, err := google.ConfigFromJSON(b, calendar.CalendarScope)
	if err != nil {
		return nil, fmt.Errorf("не удалось разобрать файл credentials: %w", err)
	}

	client := getClient(config)
	srv, err := calendar.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		return nil, fmt.Errorf("не удалось инициализировать сервис календаря: %w", err)
	}

	return &GoogleCalendarService{
		client:     srv,
		calendarID: "primary",
	}, nil
}

func (s *GoogleCalendarService) AddAppointment(appointment *common.Appointment, client *common.Client) (string, error) {
	startTime := appointment.StartTime.Add(-5 * time.Hour)
	endTime := appointment.EndTime.Add(-5 * time.Hour)

	event := &calendar.Event{
		Summary: fmt.Sprintf("Запись на приём: %s", client.Name),
		Description: fmt.Sprintf(
			"Услуга: %s\nЦена: %.2f ₽\n\nДанные клиента:\nТелефон: %s\nTelegram: %s",
			appointment.Name,
			appointment.TotalPrice,
			client.Phone,
			client.Telegram,
		),
		Start: &calendar.EventDateTime{
			DateTime: startTime.Format(time.RFC3339),
			TimeZone: "UTC",
		},
		End: &calendar.EventDateTime{
			DateTime: endTime.Format(time.RFC3339),
			TimeZone: "UTC",
		},
		ColorId: "5",
	}

	createdEvent, err := s.client.Events.Insert(s.calendarID, event).Do()
	if err != nil {
		return "", fmt.Errorf("ошибка при создании события: %w", err)
	}

	return createdEvent.Id, nil
}

func (g *GoogleCalendarService) RemoveAppointment(eventID string) error {
	if err := g.client.Events.Delete(g.calendarID, eventID).Do(); err != nil {
		return fmt.Errorf("ошибка при удалении события: %w", err)
	}
	fmt.Printf("Событие с ID %s удалено\n", eventID)
	return nil
}

func (g *GoogleCalendarService) UpdateAppointment(eventID, summary, location string, startTime, endTime time.Time) (*calendar.Event, error) {
	event, err := g.client.Events.Get(g.calendarID, eventID).Do()
	if err != nil {
		return nil, fmt.Errorf("не удалось найти событие: %w", err)
	}

	event.Summary = summary
	event.Location = location
	event.Start = &calendar.EventDateTime{
		DateTime: startTime.Format(time.RFC3339),
		TimeZone: defaultTimeZone,
	}
	event.End = &calendar.EventDateTime{
		DateTime: endTime.Format(time.RFC3339),
		TimeZone: defaultTimeZone,
	}

	updatedEvent, err := g.client.Events.Update(g.calendarID, eventID, event).Do()
	if err != nil {
		return nil, fmt.Errorf("ошибка при обновлении события: %w", err)
	}

	fmt.Printf("Событие обновлено: %s\n", updatedEvent.HtmlLink)
	return updatedEvent, nil
}
