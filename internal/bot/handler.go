package bot

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/RudinMaxim/BarberBot.git/common"
	"github.com/RudinMaxim/BarberBot.git/helper"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/google/uuid"
)

const (
	stepSelectService = iota
	stepSelectDate
	stepSelectTime
	stepConfirmBooking
	BUFFER_MINUTES   = 5
	POSSIBLE_RECORDS = 30
)

type BookingState struct {
	Step      int
	ServiceID string
	Date      time.Time
	Time      string
}

type Handler struct {
	service       *Service
	bot           *tgbotapi.BotAPI
	bookingStates map[int64]*BookingState
}

func NewHandler(service *Service, bot *tgbotapi.BotAPI) *Handler {
	return &Handler{
		service:       service,
		bot:           bot,
		bookingStates: make(map[int64]*BookingState),
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
	} else if update.CallbackQuery != nil { // Обработка CallbackQuery
		userID = update.CallbackQuery.From.ID
		chatID = update.CallbackQuery.Message.Chat.ID
		h.handleCallbackQuery(update.CallbackQuery) // Вызов обработчика CallbackQuery
	}

	if update.Message != nil && update.Message.IsCommand() {
		h.handleCommand(update, userID, chatID)
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
	data := callbackQuery.Data
	chatID := callbackQuery.Message.Chat.ID
	userID := callbackQuery.From.ID

	parts := strings.SplitN(data, ":", 2)
	action := parts[0]

	switch action {
	case "service":
		h.handleServiceSelection(chatID, userID, parts[1])
	case "date":
		h.handleDateSelection(chatID, userID, parts[1])
	case "time":
		h.handleTimeSelection(chatID, userID, parts[1])
	case "confirm_booking":
		h.handleBookingConfirmation(chatID, userID)
	case "cancel_booking":
		h.handleBookingCancellation(chatID, userID)
	case "back_to_services":
		h.bookingStates[userID].Step = stepSelectService
		h.sendServiceSelection(chatID)
	case "back_to_dates":
		h.bookingStates[userID].Step = stepSelectDate
		h.sendDateSelection(chatID)
	default:
		log.Printf("Unknown callback action: %s", action)
	}
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
	chatID := update.Message.Chat.ID
	userID := update.Message.From.ID

	client, err := h.service.GetClientByTelegramID(userID)
	if err != nil {
		log.Printf("Error getting client: %v", err)
		h.sendMessage(chatID, "Произошла ошибка при получении информации о клиенте.")
		return
	}

	if client == nil {
		h.sendMessage(chatID, "Пожалуйста, сначала зарегистрируйтесь, отправив свой контакт.")
		return
	}

	// Начинаем процесс бронирования
	h.bookingStates[userID] = &BookingState{Step: stepSelectService}
	h.sendServiceSelection(chatID)
}

func (h *Handler) sendServiceSelection(chatID int64) {
	services, err := h.service.GetActiveServices()
	if err != nil {
		log.Printf("Error getting services: %v", err)

		h.sendMessage(chatID, "Произошла ошибка при получении списка услуг.")
	}

	var keyboard [][]tgbotapi.InlineKeyboardButton
	for _, service := range services {
		button := tgbotapi.NewInlineKeyboardButtonData(
			fmt.Sprintf("%s (%d мин, %.2f руб)", service.Name, service.Duration, service.Price),
			fmt.Sprintf("service:%s", service.UUID),
		)
		keyboard = append(keyboard, []tgbotapi.InlineKeyboardButton{button})
	}
	keyboard = append(keyboard, []tgbotapi.InlineKeyboardButton{
		tgbotapi.NewInlineKeyboardButtonData("Отмена", "cancel_booking"),
	})

	msg := tgbotapi.NewMessage(chatID, "Выберите услугу:")
	msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(keyboard...)
	h.bot.Send(msg)
}

func (h *Handler) sendDateSelection(chatID int64) {
	availableDates, err := h.service.GetAvailableDates()
	if err != nil {
		log.Printf("Error getting available dates: %v", err)
		h.sendMessage(chatID, "Произошла ошибка при получении доступных дат.")
	}

	var keyboard [][]tgbotapi.InlineKeyboardButton
	for _, date := range availableDates {
		button := tgbotapi.NewInlineKeyboardButtonData(
			date.Format("02.01.2006"),
			fmt.Sprintf("date:%s", date.Format("2006-01-02")),
		)
		keyboard = append(keyboard, []tgbotapi.InlineKeyboardButton{button})
	}
	keyboard = append(keyboard, []tgbotapi.InlineKeyboardButton{
		tgbotapi.NewInlineKeyboardButtonData("Назад", "back_to_services"),
		tgbotapi.NewInlineKeyboardButtonData("Отмена", "cancel_booking"),
	})

	msg := tgbotapi.NewMessage(chatID, "Выберите дату:")
	msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(keyboard...)
	h.bot.Send(msg)
}

func (h *Handler) sendTimeSelection(chatID int64, userID int64) {
	state := h.bookingStates[userID]
	serviceIDs := []uuid.UUID{uuid.MustParse(state.ServiceID)}

	availableSlots, err := h.service.GetAvailableSlots(serviceIDs, state.Date)
	if err != nil {
		log.Printf("Error getting available time slots: %v", err)
		h.sendMessage(chatID, "Произошла ошибка при получении доступных временных слотов.")
	}

	var keyboard [][]tgbotapi.InlineKeyboardButton
	for _, slot := range availableSlots {
		button := tgbotapi.NewInlineKeyboardButtonData(
			slot.Format("15:04"),
			fmt.Sprintf("time:%s", slot.Format("15:04")),
		)
		keyboard = append(keyboard, []tgbotapi.InlineKeyboardButton{button})
	}
	keyboard = append(keyboard, []tgbotapi.InlineKeyboardButton{
		tgbotapi.NewInlineKeyboardButtonData("Назад", "back_to_dates"),
		tgbotapi.NewInlineKeyboardButtonData("Отмена", "cancel_booking"),
	})

	msg := tgbotapi.NewMessage(chatID, "Выберите время:")
	msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(keyboard...)
	h.bot.Send(msg)
}

func (h *Handler) sendBookingConfirmation(chatID int64, userID int64) {
	state := h.bookingStates[userID]
	serviceID, err := uuid.Parse(state.ServiceID)
	if err != nil {
		log.Printf("Error parsing serviceID: %v", err)
		h.sendMessage(chatID, "Некорректный идентификатор услуги.")
	}

	fmt.Println("serviceID", serviceID)

	service, err := h.service.GetServiceByID(serviceID)
	if err != nil {
		log.Printf("Error getting service: %v", err)
		h.sendMessage(chatID, "Произошла ошибка при получении информации об услуге.")
	}

	fmt.Println("service", service)

	confirmationText := fmt.Sprintf(
		"Пожалуйста, подтвердите ваше бронирование:\n\n"+
			"Услуга: %s\n"+
			"Дата: %s\n"+
			"Время: %s\n"+
			"Стоимость: %.2f руб.",
		service.Name,
		state.Date.Format("02.01.2006"),
		state.Time,
		service.Price,
	)

	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Подтвердить", "confirm_booking"),
			tgbotapi.NewInlineKeyboardButtonData("Отмена", "cancel_booking"),
		),
	)

	msg := tgbotapi.NewMessage(chatID, confirmationText)
	msg.ReplyMarkup = keyboard
	h.bot.Send(msg)
}

func (h *Handler) handleServiceSelection(chatID int64, userID int64, serviceID string) {
	err := h.service.SaveSelectedService(userID, serviceID)
	if err != nil {
		log.Printf("Error saving selected service: %v", err)
		h.sendMessage(chatID, "Произошла ошибка при выборе услуги.")
		return
	}

	h.bookingStates[userID].ServiceID = serviceID
	h.bookingStates[userID].Step = stepSelectDate
	h.sendDateSelection(chatID)
}

func (h *Handler) handleDateSelection(chatID int64, userID int64, dateStr string) {
	date, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		log.Printf("Error parsing date: %v", err)
		h.sendMessage(chatID, "Произошла ошибка при обработке выбранной даты.")
		return
	}

	err = h.service.SaveSelectedDate(userID, date)
	if err != nil {
		log.Printf("Error saving selected date: %v", err)
		h.sendMessage(chatID, "Произошла ошибка при выборе даты.")
		return
	}

	h.bookingStates[userID].Date = date
	h.bookingStates[userID].Step = stepSelectTime
	h.sendTimeSelection(chatID, userID)
}

func (h *Handler) handleTimeSelection(chatID int64, userID int64, timeStr string) {
	h.bookingStates[userID].Time = timeStr
	h.bookingStates[userID].Step = stepConfirmBooking
	h.sendBookingConfirmation(chatID, userID)
}

func (h *Handler) handleBookingConfirmation(chatID int64, userID int64) {
	state := h.bookingStates[userID]
	appointment, err := h.service.CreateAppointment(userID, state.Time)
	if err != nil {
		log.Printf("Error creating appointment: %v", err)
		h.sendMessage(chatID, "Произошла ошибка при создании записи. Пожалуйста, попробуйте еще раз.")
	}

	successMessage := fmt.Sprintf(
		"Ваша запись успешно создана!\n\n"+
			"Услуга: %s\n"+
			"Дата: %s\n"+
			"Время: %s\n"+
			"Стоимость: %.2f руб.",
		appointment.Name,
		appointment.StartTime.Format("02.01.2006"),
		appointment.StartTime.Format("15:04"),
		appointment.TotalPrice,
	)
	h.sendMessage(chatID, successMessage)

	// Очистка состояния бронирования
	delete(h.bookingStates, userID)
}

func (h *Handler) handleBookingCancellation(chatID int64, userID int64) {
	delete(h.bookingStates, userID)
	h.sendMessage(chatID, "Бронирование отменено. Если хотите начать заново, используйте команду /book.")
}

// ==================================
