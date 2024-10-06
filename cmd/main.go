package main

import (
	"log"

	"github.com/RudinMaxim/BarberBot.git/config"
	"github.com/RudinMaxim/BarberBot.git/internal/bot"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/spf13/viper"
)

type application struct {
	telegramToken string
}

// Напоминания о записи
// Отправка отзывов
// Получение информации о мастере

func main() {
	config.Init()

	telegramToken := viper.GetString("TELEGRAM_TOKEN")
	// credentialsFile := viper.GetString("GOOGLE_CREDENTIALS_FILE")
	// spreadsheetID := viper.GetString("GOOGLE_SPREADSHEET_ID")

	tgBot, err := tgbotapi.NewBotAPI(telegramToken)
	if err != nil {
		log.Fatalf("Failed to create bot: %v", err)
	}

	config.LogAction("Бот запущен")

	// ctx := context.Background()

	// sheetsRepo := sheets.NewRepository(ctx)
	// sheetsService := sheets.NewService(sheetsRepo)
	// sheetsHandler, err := sheets.NewHandler(ctx, sheetsService, credentialsFile, spreadsheetID)
	// if err != nil {
	// 	log.Fatalf("Failed to create sheets handler: %v", err)
	// }

	botRepo := bot.NewRepository()
	botService := bot.NewService(botRepo)
	botHandler := bot.NewHandler(botService, tgBot)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := tgBot.GetUpdatesChan(u)

	for update := range updates {
		if update.Message != nil {
			botHandler.HandleUpdate(update)
		}
	}
}
