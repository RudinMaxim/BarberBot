package bot

import (
	"log"
	"time"

	"github.com/RudinMaxim/BarberBot.git/common"
	"github.com/RudinMaxim/BarberBot.git/helper"
	"github.com/RudinMaxim/BarberBot.git/internal/appointments"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type Handler struct {
	service      *Service
	bot          *tgbotapi.BotAPI
	appointments appointments.Service
}

func NewHandler(service *Service, bot *tgbotapi.BotAPI, appointments appointments.Service) *Handler {
	return &Handler{
		service:      service,
		bot:          bot,
		appointments: appointments,
	}
}

// ================Common==================

func (h *Handler) HandleUpdate(update tgbotapi.Update) {
	if update.Message == nil && update.CallbackQuery == nil {
		return
	}

	var userID int64
	var chatID int64

	if update.Message != nil {
		userID = update.Message.From.ID
		chatID = update.Message.Chat.ID

		if update.Message.Contact != nil {
			h.handleContact(update.Message)
			return
		}
	} else if update.CallbackQuery != nil {
		userID = update.CallbackQuery.From.ID
		chatID = update.CallbackQuery.Message.Chat.ID
	}

	if update.Message != nil && update.Message.IsCommand() {
		h.handleCommand(update, userID, chatID)
	} else if update.CallbackQuery != nil {
		h.handleCallbackQuery(update.CallbackQuery)
	}
}

func (h *Handler) handleCommand(update tgbotapi.Update, userID int64, chatID int64) {
	switch update.Message.Command() {
	case "start":
		h.handleStart(update)
	case "help":
		h.handleHelp(update)
	case "contact":
		h.handleContactMaster(update)
	case "services":
		h.handleServices(update)
	case "about":
		h.handleAbout(update)
	case "book":
		h.handleBook(update)
	case "my_appointments":
		h.handleMyAppointments(update)
	case "cancel":
		h.handleCancel(update)
	case "reschedule":
		h.handleReschedule(update)
	default:
		h.handleUnknownCommand(update)
	}
}

func (h *Handler) handleCallbackQuery(callbackQuery *tgbotapi.CallbackQuery) {
	// TODO: Implement callback query handling if needed
}

func (h *Handler) sendMessage(chatID int64, text string) {
	msg := tgbotapi.NewMessage(chatID, text)
	_, err := h.bot.Send(msg)
	if err != nil {
		log.Printf("Error sending message: %v", err)
	}
}

// ================Static==================

func (h *Handler) sendWelcomeMessage(chatID int64) {
	h.sendMessage(chatID, helper.GetText("welcome_message"))
}

func (h *Handler) handleUnknownCommand(update tgbotapi.Update) {
	unknownCommandText := helper.GetText("unknown_command")
	msg := tgbotapi.NewMessage(update.Message.Chat.ID, unknownCommandText)
	h.bot.Send(msg)
}

func (h *Handler) handleHelp(update tgbotapi.Update) {
	h.sendMessage(update.Message.Chat.ID, helper.GetText("help_message"))
}

func (h *Handler) handleContactMaster(update tgbotapi.Update) {
	h.sendMessage(update.Message.Chat.ID, helper.GetText("contact_info"))
}

func (h *Handler) handleServices(update tgbotapi.Update) {
	h.sendMessage(update.Message.Chat.ID, helper.GetText("services_list"))
}

func (h *Handler) handleAbout(update tgbotapi.Update) {
	h.sendMessage(update.Message.Chat.ID, helper.GetText("about_master"))
}

func (h *Handler) handleMyAppointments(update tgbotapi.Update) {
	// TODO: Implement appointments viewing logic
	h.sendMessage(update.Message.Chat.ID, "Функция просмотра записей будет доступна в ближайшее время.")
}

func (h *Handler) handleCancel(update tgbotapi.Update) {
	// TODO: Implement cancellation logic
	h.sendMessage(update.Message.Chat.ID, "Функция отмены записи будет доступна в ближайшее время.")
}

func (h *Handler) handleReschedule(update tgbotapi.Update) {
	// TODO: Implement rescheduling logic
	h.sendMessage(update.Message.Chat.ID, "Функция переноса записи будет доступна в ближайшее время.")
}

// ================Dynamic==================

// Start
func (h *Handler) handleStart(update tgbotapi.Update) {
	chatID := update.Message.Chat.ID
	userID := update.Message.From.ID

	client, err := h.service.GetClientByTelegramID(userID)
	if err != nil {
		log.Printf("Error getting client: %v", err)
	}

	if client == nil {
		h.requestContact(chatID)
		return
	}

	h.sendMessage(chatID, helper.GetFormattedMessage("hello_user", update.Message.From.FirstName))
	h.sendWelcomeMessage(chatID)
}

func (h *Handler) handleContact(message *tgbotapi.Message) {
	if message.Contact == nil {
		h.sendMessage(message.Chat.ID, helper.GetText("registration_start"))
		return
	}

	username := message.From.UserName
	phone := message.Contact.PhoneNumber

	client, err := h.service.CreateClient(&common.Client{
		TelegramID: message.From.ID,
		Phone:      phone,
		Telegram:   username,
		Name:       message.From.FirstName + " " + message.From.LastName,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
		IsActive:   true,
	})

	if client == nil && err != nil {
		log.Printf("Error creating new client: %v", err)
		h.sendMessage(message.Chat.ID, helper.GetText("invalid_create_user"))
		return
	}

	h.sendMessage(message.Chat.ID, helper.GetFormattedMessage("registration_complete", message.From.FirstName))
	h.sendWelcomeMessage(message.Chat.ID)
}

func (h *Handler) requestContact(chatID int64) {
	msg := tgbotapi.NewMessage(chatID, helper.GetText("registration_start"))
	keyboard := tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButtonContact(helper.GetText("sheared_contact")),
		),
	)

	keyboard.OneTimeKeyboard = true
	msg.ReplyMarkup = keyboard

	h.bot.Send(msg)
}

// Book
func (h *Handler) handleBook(update tgbotapi.Update) {
	// TODO: Implement booking logic
	h.sendMessage(update.Message.Chat.ID, "Функция бронирования будет доступна в ближайшее время.")
}

// ==================================
