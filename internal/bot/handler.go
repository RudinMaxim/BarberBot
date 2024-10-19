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

	h.setupMenu()

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
	case "Home":
		h.handleHome(update)
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
	case "appointment":
		h.handleAppointmentSelection(chatID, parts[1])
	case "cancel":
		h.handleAppointmentCancellation(chatID, userID, parts[1])
	case "back_to_appointments":
		h.handleMyAppointments(tgbotapi.Update{Message: &tgbotapi.Message{Chat: &tgbotapi.Chat{ID: chatID}, From: &tgbotapi.User{ID: userID}}})

	default:
		h.handleUnknownCommand(tgbotapi.Update{Message: &tgbotapi.Message{Chat: &tgbotapi.Chat{ID: chatID}, From: &tgbotapi.User{ID: userID}}})
	}
}

func (h *Handler) setupMenu() {
	commands := []tgbotapi.BotCommand{
		{Command: "start", Description: "Начать работу с ботом"},
		{Command: "help", Description: "Помощь"},
		{Command: "services", Description: "Услуги"},
		{Command: "book", Description: "Записаться"},
		{Command: "my_appointments", Description: "Мои записи"},
		{Command: "cancel", Description: "Отменить запись"},
		{Command: "reschedule", Description: "Перенести запись"},
		{Command: "about", Description: "О мне"},
		{Command: "contact_master", Description: "Контакты мастера"},
	}

	cmdCfg := tgbotapi.NewSetMyCommands(
		commands...,
	)

	h.bot.Send(cmdCfg)
}

func (h *Handler) sendMessage(chatID int64, text string) {
	msg := tgbotapi.NewMessage(chatID, text)
	_, err := h.bot.Send(msg)
	if err != nil {
		log.Printf("Error sending message: %v", err)
	}
}

// ================Static==================

func (h *Handler) sendWelcomeMessage(update tgbotapi.Update) {
	h.sendMessage(update.Message.Chat.ID, helper.GetText("welcome_message"))
}

func (h *Handler) handleUnknownCommand(update tgbotapi.Update) {
	h.sendMessage(update.Message.Chat.ID, helper.GetText("unknown_command"))
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

func (h *Handler) handleHome(update tgbotapi.Update) {
	h.sendMessage(update.Message.Chat.ID, helper.GetText("home_message"))
}

// ================Dynamic==================

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
	h.sendWelcomeMessage(update)
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
	h.sendWelcomeMessage(update)
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
	availableDates, err := h.service.GetWorkingHoursAvailableDates()
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

	availableSlots, err := h.service.GetWorkingHoursAvailableSlots(serviceIDs, state.Date)
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

	service, err := h.service.GetServiceByID(serviceID)
	if err != nil {
		log.Printf("Error getting service: %v", err)
		h.sendMessage(chatID, "Произошла ошибка при получении информации об услуге.")
	}

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

// ==================================

func (h *Handler) handleMyAppointments(update tgbotapi.Update) {
	chatID := update.Message.Chat.ID
	userID := update.Message.From.ID

	appointments, err := h.service.GetClientAppointments(userID)
	if err != nil {
		log.Printf("Error getting appointments: %v", err)
		h.sendMessage(chatID, "Не удалось получить ваши записи. Попробуйте позже.")
		return
	}

	if len(appointments) == 0 {
		h.sendMessage(chatID, "У вас нет активных записей. Используйте /book для записи.")
		return
	}

	h.sendAppointmentsList(chatID, appointments)
}

func (h *Handler) sendAppointmentsList(chatID int64, appointments []common.Appointment) {
	const appointmentsPerPage = 5
	totalPages := (len(appointments) + appointmentsPerPage - 1) / appointmentsPerPage

	var keyboard [][]tgbotapi.InlineKeyboardButton

	for i, appointment := range appointments {
		if i%appointmentsPerPage == 0 {
			keyboard = append(keyboard, []tgbotapi.InlineKeyboardButton{})
		}
		buttonText := fmt.Sprintf("%s - %s", appointment.StartTime.Format("02.01 15:04"), appointment.Name)
		callbackData := fmt.Sprintf("appointment:%s", appointment.UUID)
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

	// TODO: Добавить кнопку "Новая запись"
	// keyboard = append(keyboard, []tgbotapi.InlineKeyboardButton{
	// 	tgbotapi.NewInlineKeyboardButtonData("Новая запись", ""),
	// })

	msg := tgbotapi.NewMessage(chatID, "Все ваши записи:")
	msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(keyboard...)
	h.bot.Send(msg)
}

func (h *Handler) handleAppointmentSelection(chatID int64, appointmentID string) {
	appointment, err := h.service.GetAppointmentByID(uuid.MustParse(appointmentID))
	if err != nil {
		log.Printf("Error getting appointment: %v", err)
		h.sendMessage(chatID, "Не удалось получить информацию о записи.")
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
			tgbotapi.NewInlineKeyboardButtonData("Отменить запись", fmt.Sprintf("cancel:%s", appointmentID)),
			tgbotapi.NewInlineKeyboardButtonData("Назад", "back_to_appointments"),
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
		h.sendMessage(chatID, "Не удалось получить ваши записи. Попробуйте позже.")
		return
	}

	// Если записей нет
	if len(appointments) == 0 {
		h.sendMessage(chatID, "У вас нет активных записей для отмены.")
		return
	}

	// Формируем список записей для выбора
	var keyboard [][]tgbotapi.InlineKeyboardButton
	for _, appointment := range appointments {
		buttonText := fmt.Sprintf("%s - %s", appointment.StartTime.Format("02.01 15:04"), appointment.Name)
		callbackData := fmt.Sprintf("cancel:%s", appointment.UUID)
		button := tgbotapi.NewInlineKeyboardButtonData(buttonText, callbackData)
		keyboard = append(keyboard, []tgbotapi.InlineKeyboardButton{button})
	}

	// Добавляем кнопку для выхода
	keyboard = append(keyboard, []tgbotapi.InlineKeyboardButton{
		tgbotapi.NewInlineKeyboardButtonData("Назад", "back_to_appointments"),
	})

	// Отправляем сообщение с выбором записи для отмены
	msg := tgbotapi.NewMessage(chatID, "Выберите запись для отмены:")
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

	h.sendMessage(chatID, "Ваша запись успешно отменена.")
	h.handleMyAppointments(tgbotapi.Update{Message: &tgbotapi.Message{Chat: &tgbotapi.Chat{ID: chatID}, From: &tgbotapi.User{ID: userID}}})
}

func (h *Handler) handleBookingCancellation(chatID int64, userID int64) {
	delete(h.bookingStates, userID)
	h.sendMessage(chatID, "Бронирование отменено. Если хотите начать заново, используйте команду /book.")
}

// ==================================

func (h *Handler) handleReschedule(update tgbotapi.Update) {
	// TODO: Реализовать функцию переноса записи
	h.sendMessage(update.Message.Chat.ID, "Функция переноса записи будет доступна в ближайшее время.")
}
