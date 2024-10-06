package bot

import (
	"bytes"
	"fmt"
	"html/template"
	"log"
	"strings"

	"github.com/RudinMaxim/BarberBot.git/config"
	"github.com/RudinMaxim/BarberBot.git/helper"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type Handler struct {
	service *Service
	bot     *tgbotapi.BotAPI
}

func NewHandler(service *Service, bot *tgbotapi.BotAPI) *Handler {
	return &Handler{
		service: service,
		bot:     bot,
	}
}

func (h *Handler) HandleUpdate(update tgbotapi.Update) {
	if update.Message == nil {
		return
	}

	userID := update.Message.From.ID
	chatID := update.Message.Chat.ID

	registrationState := h.service.GetRegistrationState(userID)
	if registrationState.State != StateNone {
		// Если пользователь в процессе регистрации, обрабатываем ввод
		h.handleRegister(update)
		return
	}

	if update.Message != nil {
		switch update.Message.Command() {
		case "start":
			h.handleStart(update) // done
		case "help":
			h.handleHelp(update) // done
		case "contact":
			h.handleContact(update) // done
		case "services":
			h.handleServices(update) // done
		case "about":
			h.handleAbout(update) // done
		case "register":
			h.handleRegister(update)
		case "book":
			// TODO: Implement reschedule command
		case "reschedule":
			// TODO: Implement reschedule command
		case "my_appointments":
			// TODO: Implement reschedule command
		case "cancel":
			// TODO: Implement reschedule command
		default:
			h.handleUnknownCommand(update)
		}
	} else {
		// Если это не команда и пользователь не в процессе регистрации,
		// отправляем сообщение о неизвестной команде
		msg := tgbotapi.NewMessage(chatID, "Извините, я не понимаю это сообщение. Пожалуйста, используйте команды для взаимодействия со мной.")
		h.bot.Send(msg)
	}
}

func (h *Handler) handleStart(update tgbotapi.Update) {
	username := update.Message.From.UserName
	if username == "" {
		username = "дорогой клиент"
	}
	welcomeTemplate := helper.GetText("welcome_message")

	// Используем шаблон и подставляем данные
	tmpl, err := template.New("welcome").Parse(welcomeTemplate)
	if err != nil {
		log.Fatalf("Error parsing welcome message template: %v", err)
	}

	var msgBuffer bytes.Buffer
	tmpl.Execute(&msgBuffer, map[string]string{
		"Username": username,
	})

	msg := tgbotapi.NewMessage(update.Message.Chat.ID, msgBuffer.String())
	h.bot.Send(msg)
}

func (h *Handler) handleUnknownCommand(update tgbotapi.Update) {
	unknownCommandText := helper.GetText("unknown_command")
	msg := tgbotapi.NewMessage(update.Message.Chat.ID, unknownCommandText)
	h.bot.Send(msg)
}

func (h *Handler) handleHelp(update tgbotapi.Update) {
	helpText := helper.GetText("help")
	msg := tgbotapi.NewMessage(update.Message.Chat.ID, helpText)
	h.bot.Send(msg)
}

func (h *Handler) handleContact(update tgbotapi.Update) {
	contactText := helper.GetText("contact")
	msg := tgbotapi.NewMessage(update.Message.Chat.ID, contactText)
	h.bot.Send(msg)
}

func (h *Handler) handleServices(update tgbotapi.Update) {
	servicesText := helper.GetText("services")
	msg := tgbotapi.NewMessage(update.Message.Chat.ID, servicesText)
	h.bot.Send(msg)
}

func (h *Handler) handleAbout(update tgbotapi.Update) {
	aboutText := helper.GetText("about_master")
	msg := tgbotapi.NewMessage(update.Message.Chat.ID, aboutText)
	h.bot.Send(msg)
}

func (h *Handler) handleRegister(update tgbotapi.Update) {
	userID := update.Message.From.ID
	chatID := update.Message.Chat.ID
	text := update.Message.Text

	result := h.service.HandleRegistration(userID, text)

	msg := tgbotapi.NewMessage(chatID, result.Message)
	h.bot.Send(msg)

	if result.Done {
		fmt.Printf("User registered: %d\n", userID)
		config.LogAction("User registered")
	}
}

// Helper to split command arguments
func splitArguments(args string) []string {
	return strings.Fields(args)
}
