package bot

import (
	"fmt"
	"log"
	"strconv"
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

var commands = []tgbotapi.BotCommand{
	{Command: "home", Description: "На главную"},
	{Command: "location", Description: "Наша локация"},
	{Command: "help", Description: "Помощь"},
	{Command: "services", Description: "Услуги"},
	{Command: "book", Description: "Записаться"},
	{Command: "my_appointments", Description: "Мои записи"},
	{Command: "cancel", Description: "Отменить запись"},
	{Command: "reschedule", Description: "Перенести запись"},
	{Command: "about", Description: "О мне"},
	{Command: "contact", Description: "Контакты мастера"},
}

type BookingState struct {
	Step          int
	ServiceID     string
	Date          time.Time
	Time          string
	AppointmentID string
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

	resp, err := h.bot.Request(tgbotapi.NewSetMyCommands(commands...))
	if err != nil {
		log.Printf("Error setting commands: %v", err)
		return
	}

	if resp.Ok {
		log.Println("Commands were set successfully.")
	} else {
		log.Printf("Failed to set commands: %s", resp.Description)
	}

	if update.Message != nil {
		if update.Message.Contact != nil {
			h.handleContact(update)
			return
		}
	} else if update.CallbackQuery != nil {
		h.handleCallbackQuery(update.CallbackQuery)
	}

	if update.Message != nil && update.Message.IsCommand() {
		h.handleCommand(update)
	}
}

func (h *Handler) handleCommand(update tgbotapi.Update) {
	switch update.Message.Command() {
	case "start":
		h.handleStart(update)
	case "location":
		h.handleLocation(update)
	case "home":
		h.handleHome(update)
	case "consultation":
		h.handleConsultation(update)
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
	chatID := callbackQuery.Message.Chat.ID
	userID := callbackQuery.From.ID
	data := callbackQuery.Data

	// Обработка простых действий без параметров
	switch data {
	case "confirm_booking":
		h.handleBookingConfirmation(chatID, userID)
		return
	case "go_home":
		h.handleHome(tgbotapi.Update{Message: &tgbotapi.Message{Chat: &tgbotapi.Chat{ID: chatID}, From: &tgbotapi.User{ID: userID}}})
		return
	case "back_to_services":
		h.bookingStates[userID].Step = stepSelectService
		h.sendServiceSelection(chatID)
		return
	case "back_to_dates":
		h.bookingStates[userID].Step = stepSelectDate
		h.sendDateSelection(chatID)
		return
	case "back_to_appointments":
		h.handleMyAppointments(tgbotapi.Update{Message: &tgbotapi.Message{Chat: &tgbotapi.Chat{ID: chatID}, From: &tgbotapi.User{ID: userID}}})
		return
	case "new_appointment":
		h.handleBook(tgbotapi.Update{Message: &tgbotapi.Message{Chat: &tgbotapi.Chat{ID: chatID}, From: &tgbotapi.User{ID: userID}}})
		return
	}

	// Обработка действий с параметрами
	parts := strings.SplitN(data, ":", 2)
	if len(parts) != 2 {
		log.Printf("Invalid callback data format: %s", data)
		h.sendMessage(chatID, "Произошла ошибка при обработке команды")
		return
	}

	action := parts[0]
	value := parts[1]

	switch action {
	case "service":
		h.handleServiceSelection(chatID, userID, value)
	case "date":
		h.handleDateSelection(chatID, userID, value)
	case "time":
		if h.bookingStates[userID].AppointmentID != "" {
			h.handleRescheduleTimeSelection(chatID, userID, value)
		} else {
			h.handleTimeSelection(chatID, userID, value)
		}
	case "appointment":
		h.handleAppointmentSelection(chatID, value)
	case "cancel":
		h.handleAppointmentCancellation(chatID, userID, value)
	case "reschedule":
		h.handleAppointmentReschedule(chatID, userID, value)
	case "page":
		page, err := strconv.Atoi(value)
		if err != nil {
			log.Printf("Error parsing page number: %v", err)
			return
		}
		appointments, err := h.service.GetClientAppointments(int64(userID))
		if err != nil {
			log.Printf("Error getting appointments: %v", err)
			return
		}
		h.sendAppointmentsPage(chatID, appointments, page, callbackQuery.Message.MessageID)
	default:
		log.Printf("Unknown callback action: %s", action)
		h.handleUnknownCommand(tgbotapi.Update{Message: &tgbotapi.Message{Chat: &tgbotapi.Chat{ID: chatID}, From: &tgbotapi.User{ID: userID}}})
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

func (h *Handler) handleUnknownCommand(update tgbotapi.Update) {
	h.sendMessage(update.Message.Chat.ID, helper.GetText("unknown_command"))
}

func (h *Handler) handleHelp(update tgbotapi.Update) {
	h.sendMessage(update.Message.Chat.ID, helper.GetText("help_message"))
}

func (h *Handler) handleConsultation(update tgbotapi.Update) {
	h.sendMessage(update.Message.Chat.ID, helper.GetText("consultation_message"))
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

func (h *Handler) handleHome(update tgbotapi.Update) {
	h.sendMessage(update.Message.Chat.ID, helper.GetText("home_message"))
}

func (h *Handler) handleLocation(update tgbotapi.Update) {
	msg := tgbotapi.NewMessage(update.Message.Chat.ID, helper.GetText("location_message"))

	yandexButton := tgbotapi.NewInlineKeyboardButtonURL("Открыть в Yandex Maps", "https://yandex.ru/maps/-/CDdRZLYM")

	dgisButton := tgbotapi.NewInlineKeyboardButtonURL("Открыть в 2GIS", "https://go.2gis.com/ofrhv")

	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(yandexButton, dgisButton),
	)

	msg.ReplyMarkup = keyboard

	_, err := h.bot.Send(msg)
	if err != nil {
		log.Printf("Error sending location message: %v", err)
	}
}

// ================Dynamic==================

func (h *Handler) handleStart(update tgbotapi.Update) {
	chatID := update.Message.Chat.ID
	userID := update.Message.From.ID
	firstName := update.Message.From.FirstName

	_, err := h.service.GetClientBy("telegram_id", userID)
	if err != nil {
		log.Printf("Error getting client: %v", err)
		h.requestContact(chatID)
		return
	}

	h.sendMessage(chatID, helper.GetFormattedMessage("hello_user", firstName))
	h.handleHome(update)
}

func (h *Handler) handleContact(update tgbotapi.Update) {
	message := update.Message

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
	h.handleHome(update)
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

// ==================================

func (h *Handler) handleBook(update tgbotapi.Update) {
	chatID := update.Message.Chat.ID
	userID := update.Message.From.ID

	client, err := h.service.GetClientBy("telegram_id", userID)
	if err != nil {
		log.Printf("Error getting client: %v", err)
		h.sendMessage(chatID, helper.GetText("invalid_get_user"))
	}

	if client == nil {
		h.sendMessage(chatID, helper.GetText("user_dont_registered"))
		return
	}

	h.bookingStates[userID] = &BookingState{Step: stepSelectService}
	h.sendServiceSelection(chatID)
}

func (h *Handler) sendServiceSelection(chatID int64) {
	services, err := h.service.GetActiveServices()
	if err != nil {
		log.Printf("Error getting services: %v", err)
		h.sendMessage(chatID, helper.GetText("invalid_get_services"))
		return
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
		tgbotapi.NewInlineKeyboardButtonData(helper.GetText("cancel_button"), "go_home"),
	})

	msg := tgbotapi.NewMessage(chatID, helper.GetText("select_service"))
	msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(keyboard...)
	h.bot.Send(msg)
}

func (h *Handler) sendDateSelection(chatID int64) {
	availableDates, err := h.service.GetWorkingHoursAvailableDates()
	if err != nil {
		log.Printf("Error getting available dates: %v", err)
		h.sendMessage(chatID, helper.GetText("invalid_get_dates"))
		return
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
		tgbotapi.NewInlineKeyboardButtonData(helper.GetText("back_button"), "back_to_services"),
		tgbotapi.NewInlineKeyboardButtonData(helper.GetText("cancel_button"), "go_home"),
	})

	msg := tgbotapi.NewMessage(chatID, helper.GetText("select_date"))
	msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(keyboard...)
	h.bot.Send(msg)
}

func (h *Handler) sendTimeSelection(chatID int64, userID int64) {
	state := h.bookingStates[userID]
	serviceIDs := []uuid.UUID{uuid.MustParse(state.ServiceID)}

	availableSlots, err := h.service.GetWorkingHoursAvailableSlots(serviceIDs, state.Date)
	if err != nil {
		log.Printf("Error getting available time slots: %v", err)
		h.sendMessage(chatID, helper.GetText("invalid_get_slots"))
		return
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
		tgbotapi.NewInlineKeyboardButtonData(helper.GetText("back_button"), "back_to_dates"),
		tgbotapi.NewInlineKeyboardButtonData(helper.GetText("cancel_button"), "go_home"),
	})

	msg := tgbotapi.NewMessage(chatID, helper.GetText("select_time"))
	msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(keyboard...)
	h.bot.Send(msg)
}

func (h *Handler) sendBookingConfirmation(chatID int64, userID int64) {
	state := h.bookingStates[userID]
	serviceID, err := uuid.Parse(state.ServiceID)
	if err != nil {
		log.Printf("Error parsing serviceID: %v", err)
		h.sendMessage(chatID, helper.GetText("invalid_id_service"))
		return
	}

	service, err := h.service.GetServiceByID(serviceID)
	if err != nil {
		log.Printf("Error getting service: %v", err)
		h.sendMessage(chatID, helper.GetText("invalid_id_service"))
		return
	}

	confirmationText := fmt.Sprintf(
		"Пожалуйста, подтвердите ваше бронирование:\n\n"+
			"🗓 Дата: %s\n\n"+
			"💇 Услуга: %s\n\n"+
			"🕒 Время: %s (сеанса: %d минут)\n\n"+
			"💰 Стоимость: %.2f руб.\n\n",
		state.Date.Format("02.01.2006"),
		service.Name,
		state.Time,
		service.Duration,
		service.Price,
	)

	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(helper.GetText("confirm_button"), "confirm_booking"),
			tgbotapi.NewInlineKeyboardButtonData(helper.GetText("cancel_button"), "go_home"),
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
		h.sendMessage(chatID, helper.GetText("invalid_get_services"))
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
	stateTime := h.bookingStates[userID].Time
	appointment, err := h.service.CreateAppointment(userID, stateTime)
	log.Printf("Appointment created: %+v", appointment)
	if err != nil {
		log.Printf("Error creating appointment: %v", err)
		h.sendMessage(chatID, helper.GetText("invalid_create_appointment"))
		return
	}

	successMessage := fmt.Sprintf(
		"🌟 Ура, ваша запись успешно создана!\n\n"+
			"🗓 Дата: %s\n\n"+
			"💇 Услуга: %s\n\n"+
			"🕒 Время: %s\n\n"+
			"💰 Стоимость: %.2f руб.\n\n"+
			"🚩 Адрес: улица Куйбышева, 79.",
		appointment.Name,
		appointment.StartTime.Format("02.01.2006"),
		appointment.StartTime.Format("15:04"),
		appointment.TotalPrice,
	)
	h.sendMessage(chatID, successMessage)

	delete(h.bookingStates, userID)
}

// ==================================

func (h *Handler) handleMyAppointments(update tgbotapi.Update) {
	chatID := update.Message.Chat.ID
	userID := update.Message.From.ID

	appointments, err := h.service.GetClientAppointments(userID)
	if err != nil {
		log.Printf("Error getting appointments: %v", err)
		h.sendMessage(chatID, helper.GetText("invalid_get_appointments"))
		return
	}

	if len(appointments) == 0 {
		h.sendMessage(chatID, helper.GetText("no_appointments"))
		return
	}

	h.sendAppointmentsList(chatID, appointments)
}

func (h *Handler) sendAppointmentsList(chatID int64, appointments []common.Appointment) {
	const appointmentsPerPage = 5
	totalPages := (len(appointments) + appointmentsPerPage - 1) / appointmentsPerPage

	var keyboard [][]tgbotapi.InlineKeyboardButton

	for _, appointment := range appointments[:min(appointmentsPerPage, len(appointments))] {
		buttonText := fmt.Sprintf("%s - %s", appointment.StartTime.Format("02.01 15:04"), appointment.Name)
		callbackData := fmt.Sprintf("appointment:%s", appointment.UUID)
		button := tgbotapi.NewInlineKeyboardButtonData(buttonText, callbackData)
		keyboard = append(keyboard, []tgbotapi.InlineKeyboardButton{button})
	}

	// Add navigation buttons
	var navigationRow []tgbotapi.InlineKeyboardButton
	if totalPages > 1 {
		if len(appointments) > appointmentsPerPage {
			navigationRow = append(navigationRow, tgbotapi.NewInlineKeyboardButtonData("➡️", fmt.Sprintf("page:1")))
		}
	}

	if len(navigationRow) > 0 {
		keyboard = append(keyboard, navigationRow)
	}

	msg := tgbotapi.NewMessage(chatID, helper.GetText("select_appointment"))
	msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(keyboard...)

	h.bot.Send(msg)
}

func (h *Handler) handleAppointmentsPageChange(query *tgbotapi.CallbackQuery) {
	parts := strings.SplitN(query.Data, ":", 2)
	if len(parts) != 2 {
		return
	}

	pageStr := parts[1]
	page, err := strconv.Atoi(pageStr)
	if err != nil {
		return
	}

	userID := query.From.ID
	appointments, err := h.service.GetClientAppointments(int64(userID))
	if err != nil {
		log.Printf("Error getting appointments: %v", err)
		return
	}

	h.sendAppointmentsPage(query.Message.Chat.ID, appointments, page, query.Message.MessageID)
}

func (h *Handler) sendAppointmentsPage(chatID int64, appointments []common.Appointment, page int, messageID int) {
	const appointmentsPerPage = 5
	totalPages := (len(appointments) + appointmentsPerPage - 1) / appointmentsPerPage

	startIndex := (page - 1) * appointmentsPerPage
	endIndex := min(startIndex+appointmentsPerPage, len(appointments))

	var keyboard [][]tgbotapi.InlineKeyboardButton

	for _, appointment := range appointments[startIndex:endIndex] {
		buttonText := fmt.Sprintf("%s - %s", appointment.StartTime.Format("02.01 15:04"), appointment.Name)
		callbackData := fmt.Sprintf("appointment:%s", appointment.UUID)
		button := tgbotapi.NewInlineKeyboardButtonData(buttonText, callbackData)
		keyboard = append(keyboard, []tgbotapi.InlineKeyboardButton{button})
	}

	// Add navigation buttons
	var navigationRow []tgbotapi.InlineKeyboardButton
	if page > 1 {
		navigationRow = append(navigationRow, tgbotapi.NewInlineKeyboardButtonData("⬅️", fmt.Sprintf("page:%d", page-1)))
	}
	if page < totalPages {
		navigationRow = append(navigationRow, tgbotapi.NewInlineKeyboardButtonData("➡️", fmt.Sprintf("page:%d", page+1)))
	}

	if len(navigationRow) > 0 {
		keyboard = append(keyboard, navigationRow)
	}

	msg := tgbotapi.NewEditMessageText(chatID, messageID, helper.GetText("select_appointment"))
	msg.ReplyMarkup = &tgbotapi.InlineKeyboardMarkup{InlineKeyboard: keyboard}

	_, err := h.bot.Send(msg)
	if err != nil {
		log.Printf("Error sending edited message: %v", err)
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func (h *Handler) handleAppointmentSelection(chatID int64, appointmentID string) {
	appointment, err := h.service.GetAppointmentByID(uuid.MustParse(appointmentID))
	if err != nil {
		log.Printf("Error getting appointment: %v", err)
		h.sendMessage(chatID, helper.GetText("invalid_get_appointment"))
		return
	}

	messageText := fmt.Sprintf(
		"Детали записи:\n\n"+
			"🗓 Дата: %s\n\n"+
			"🕒 Время: %s - %s\n\n"+
			"💇 Услуга: %s\n\n"+
			"💰 Стоимость: %.2f руб.\n\n"+
			"📊 Статус: %s",
		appointment.StartTime.Format("02.01.2006"),
		appointment.StartTime.Format("15:04"),
		appointment.EndTime.Format("15:04"),
		appointment.Name,
		appointment.TotalPrice,
		getStatusEmoji(appointment.Status),
	)

	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Отменить", fmt.Sprintf("cancel:%s", appointmentID)),
			tgbotapi.NewInlineKeyboardButtonData(helper.GetText("back_button"), "back_to_appointments"),
		),
	)

	msg := tgbotapi.NewMessage(chatID, messageText)
	msg.ReplyMarkup = keyboard
	h.bot.Send(msg)
}

func getStatusEmoji(status string) string {
	switch status {
	case "scheduled":
		return "✅ Запланировано"
	case "completed":
		return "✔️ Завершено"
	case "cancelled":
		return "❌ Отменено"
	default:
		return "❓ Неизвестно"
	}
}

// ==================================

func (h *Handler) handleCancel(update tgbotapi.Update) {
	chatID := update.Message.Chat.ID
	userID := update.Message.From.ID

	appointments, err := h.service.GetClientScheduledAppointmentsByID(userID)
	if err != nil {
		log.Printf("Error getting appointments: %v", err)
		h.sendMessage(chatID, helper.GetText("invalid_get_appointments"))
		return
	}

	if len(appointments) == 0 {
		h.sendMessage(chatID, helper.GetText("no_cancel_appointment"))
		return
	}

	var keyboard [][]tgbotapi.InlineKeyboardButton
	for _, appointment := range appointments {
		buttonText := fmt.Sprintf("%s - %s", appointment.StartTime.Format("02.01 15:04"), appointment.Name)
		callbackData := fmt.Sprintf("cancel:%s", appointment.UUID)
		button := tgbotapi.NewInlineKeyboardButtonData(buttonText, callbackData)
		keyboard = append(keyboard, []tgbotapi.InlineKeyboardButton{button})
	}

	keyboard = append(keyboard, []tgbotapi.InlineKeyboardButton{
		tgbotapi.NewInlineKeyboardButtonData(helper.GetText("back_button"), "back_to_appointments"),
	})

	// Отправляем сообщение с выбором записи для отмены
	msg := tgbotapi.NewMessage(chatID, helper.GetText("select_cancel_appointment"))
	msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(keyboard...)
	h.bot.Send(msg)
}

func (h *Handler) handleAppointmentCancellation(chatID int64, userID int64, appointmentID string) {
	uuid, err := uuid.Parse(appointmentID)
	if err != nil {
		h.sendMessage(chatID, "Неверный идентификатор записи.")
		return
	}

	err = h.service.CancelAppointment(userID, uuid)
	if err != nil {
		log.Printf("Error cancelling appointment: %v", err)
		h.sendMessage(chatID, "Произошла ошибка при отмене записи: "+err.Error())
		return
	}

	h.handleBookingCancellation(chatID, userID)
	h.handleMyAppointments(tgbotapi.Update{Message: &tgbotapi.Message{Chat: &tgbotapi.Chat{ID: chatID}, From: &tgbotapi.User{ID: userID}}})
}

func (h *Handler) handleBookingCancellation(chatID int64, userID int64) {
	delete(h.bookingStates, userID)
	h.sendMessage(chatID, helper.GetText("appointment_cancel"))
}

// ==================================

func (h *Handler) handleReschedule(update tgbotapi.Update) {
	chatID := update.Message.Chat.ID
	// userID := update.Message.From.ID

	h.sendMessage(chatID, " Функция находится на этапе разработки")

	// appointments, err := h.service.GetClientScheduledAppointmentsByID(userID)
	// if err != nil {
	// 	log.Printf("Error getting appointments: %v", err)
	// 	h.sendMessage(chatID, helper.GetText("invalid_get_appointments"))
	// 	return
	// }

	// if len(appointments) == 0 {
	// 	h.sendMessage(chatID, helper.GetText("no_reschedule_appointment"))
	// 	return
	// }

	// h.sendRescheduleAppointmentsList(chatID, appointments)
}

func (h *Handler) sendRescheduleAppointmentsList(chatID int64, appointments []common.Appointment) {
	const appointmentsPerPage = 5
	totalPages := (len(appointments) + appointmentsPerPage - 1) / appointmentsPerPage

	var keyboard [][]tgbotapi.InlineKeyboardButton

	for i, appointment := range appointments {
		if i%appointmentsPerPage == 0 {
			keyboard = append(keyboard, []tgbotapi.InlineKeyboardButton{})
		}
		buttonText := fmt.Sprintf("%s - %s", appointment.StartTime.Format("02.01 15:04"), appointment.Name)
		callbackData := fmt.Sprintf("reschedule:%s", appointment.UUID)
		button := tgbotapi.NewInlineKeyboardButtonData(buttonText, callbackData)
		keyboard[i/appointmentsPerPage] = append(keyboard[i/appointmentsPerPage], button)
	}

	if totalPages > 1 {
		var navigationRow []tgbotapi.InlineKeyboardButton
		for i := 0; i < totalPages; i++ {
			pageButton := tgbotapi.NewInlineKeyboardButtonData(fmt.Sprintf("%d", i+1), fmt.Sprintf("page:%d", i))
			navigationRow = append(navigationRow, pageButton)
		}
		keyboard = append(keyboard, navigationRow)
	}

	keyboard = append(keyboard, []tgbotapi.InlineKeyboardButton{
		tgbotapi.NewInlineKeyboardButtonData(helper.GetText("back_button"), "back_to_appointments"),
	})

	msg := tgbotapi.NewMessage(chatID, helper.GetText("select_reschedule_appointment"))
	msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(keyboard...)

	h.bot.Send(msg)
}

func (h *Handler) handleAppointmentReschedule(chatID int64, userID int64, appointmentID string) {
	uuid, err := uuid.Parse(appointmentID)
	if err != nil {
		h.sendMessage(chatID, "Неверный идентификатор записи.")
		return
	}

	appointment, err := h.service.GetAppointmentByID(uuid)
	if err != nil {
		log.Printf("Error getting appointment: %v", err)
		h.sendMessage(chatID, helper.GetText("invalid_get_appointment"))
		return
	}

	h.bookingStates[userID] = &BookingState{
		Step:      stepSelectDate,
		ServiceID: appointment.Services[0].UUID.String(),
	}

	h.sendDateSelection(chatID)
}

func (h *Handler) handleRescheduleTimeSelection(chatID int64, userID int64, timeStr string) {
	state := h.bookingStates[userID]
	appointmentID, err := uuid.Parse(state.AppointmentID)
	if err != nil {
		h.sendMessage(chatID, "Ошибка при обработке идентификатора записи.")
		return
	}

	err = h.service.RescheduleAppointment(userID, appointmentID, state.Date, timeStr)
	if err != nil {
		log.Printf("Error rescheduling appointment: %v", err)
		h.sendMessage(chatID, "Произошла ошибка при переносе записи: "+err.Error())
		return
	}

	h.sendMessage(chatID, helper.GetText("appointment_rescheduled"))
	delete(h.bookingStates, userID)
	h.handleMyAppointments(tgbotapi.Update{Message: &tgbotapi.Message{Chat: &tgbotapi.Chat{ID: chatID}, From: &tgbotapi.User{ID: userID}}})
}
