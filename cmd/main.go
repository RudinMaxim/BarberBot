package main

import (
	"fmt"
	"log"

	"github.com/RudinMaxim/BarberBot.git/config"
	"github.com/RudinMaxim/BarberBot.git/database"
	"github.com/RudinMaxim/BarberBot.git/internal/bot"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"gorm.io/gorm"
)

type application struct {
	db  *gorm.DB
	bot *tgbotapi.BotAPI
}

func main() {
	app := &application{}

	config.LogAction("Initializing application...")
	if err := app.initialize(); err != nil {
		config.LogAction(fmt.Sprintf("Failed to initialize application: %v", err))
		log.Fatalf("Failed to initialize application: %v", err)
	}

	botRepo := bot.NewRepository(app.db)
	botService := bot.NewClientService(botRepo)
	botHandler := bot.NewHandler(botService, app.bot)

	config.LogAction("Bot components created")
	config.LogAction("Bot started")

	app.runBot(botHandler)
}

func (app *application) initialize() error {
	config.LogAction("Initializing configuration...")
	config.Init()
	config.LogAction("Configuration initialized")

	config.LogAction("Initializing database...")
	if err := app.initDatabase(); err != nil {
		return fmt.Errorf("failed to initialize database: %w", err)
	}
	config.LogAction("Database initialized")

	config.LogAction("Initializing bot...")
	if err := app.initBot(); err != nil {
		return fmt.Errorf("failed to initialize bot: %w", err)
	}
	config.LogAction("Bot initialized")

	return nil
}

func (app *application) initDatabase() error {
	db, err := database.InitDatabase()
	if err != nil {
		return fmt.Errorf("could not initialize database connection: %w", err)
	}

	if err := database.PingDatabase(db); err != nil {
		return fmt.Errorf("could not ping database: %w", err)
	}

	app.db = db
	return nil
}

func (app *application) initBot() error {
	config.LogAction("Creating new BotAPI instance...")
	bot, err := tgbotapi.NewBotAPI("8008874726:AAGyfIFIPUDnmDDN0Cr5xjCF84Cf0NzqYLs")
	if err != nil {
		config.LogAction(fmt.Sprintf("Failed to create BotAPI: %v", err))
		return err
	}
	config.LogAction("BotAPI instance created successfully")

	app.bot = bot
	return nil
}

func (app *application) runBot(handler *bot.Handler) {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	config.LogAction("Starting to receive updates...")
	updates := app.bot.GetUpdatesChan(u)

	for update := range updates {
		if update.Message != nil {
			config.LogAction(fmt.Sprintf("Received message from user %d: %s", update.Message.From.ID, update.Message.Text))
			handler.HandleUpdate(update)
		}
	}
}
