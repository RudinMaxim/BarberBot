package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/RudinMaxim/BarberBot.git/config"
	"github.com/RudinMaxim/BarberBot.git/database"
	"github.com/RudinMaxim/BarberBot.git/internal/bot"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"gorm.io/gorm"
)

type application struct {
	db     *gorm.DB
	bot    *tgbotapi.BotAPI
	cache  *database.RedisCache
	ctx    context.Context
	cancel context.CancelFunc
}

func main() {
	app := &application{}

	config.LogAction("Initializing application...")
	if err := app.initialize(); err != nil {
		config.LogAction(fmt.Sprintf("Failed to initialize application: %v", err))
		log.Fatalf("Failed to initialize application: %v", err)
	}

	botRepo := bot.NewRepository(app.db, app.cache)
	botService := bot.NewService(botRepo)
	botHandler := bot.NewHandler(botService, app.bot)

	config.LogAction("Bot components created")
	config.LogAction("Bot started")

	app.runBot(botHandler)
}

func (app *application) initialize() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	app.ctx = ctx
	app.cancel = cancel

	config.LogAction("Initializing configuration...")
	config.Init()
	config.LogAction("Configuration initialized")

	config.LogAction("Initializing database...")
	if err := app.initDatabase(); err != nil {
		return fmt.Errorf("failed to initialize database: %w", err)
	}
	config.LogAction("Database initialized")

	config.LogAction("Initializing cache...")
	if err := app.initCache(); err != nil {
		return fmt.Errorf("failed to initialize cache: %w", err)
	}
	config.LogAction("Cache initialized")

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

func (app *application) initCache() error {
	redisCache := database.NewRedisCache("redis:6379")

	if err := redisCache.Ping(app.ctx); err != nil {
		return fmt.Errorf("failed to ping Redis: %w", err)
	}

	app.cache = redisCache
	return nil
}

func (app *application) runBot(handler *bot.Handler) {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 180

	config.LogAction("Starting to receive updates...")
	updates := app.bot.GetUpdatesChan(u)

	for update := range updates {
		if update.Message != nil {
			config.LogAction(fmt.Sprintf("Received message from user %d: %s", update.Message.From.ID, update.Message.Text))
			handler.HandleUpdate(update)
		} else if update.CallbackQuery != nil {
			config.LogAction(fmt.Sprintf("Received callback query from user %d: %s", update.CallbackQuery.From.ID, update.CallbackQuery.Data))
			handler.HandleUpdate(update)
		}
	}
}
