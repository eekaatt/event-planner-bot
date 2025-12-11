package bot

import (
    "fmt"
    "log"
    "strconv"
    "strings"
    "time"

    "github.com/eekaatt/event_planner_bot-go/internal/auth"
    "github.com/eekaatt/event_planner_bot-go/internal/database"
    "github.com/eekaatt/event_planner_bot-go/internal/models"
    tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type BotHandler struct {
    bot    *tgbotapi.BotAPI
    repo   *database.Repository
    auth   *auth.AuthService
    adminID int64
}

func NewBotHandler(bot *tgbotapi.BotAPI, repo *database.Repository, auth *auth.AuthService, adminID int64) *BotHandler {
    return &BotHandler{
        bot:     bot,
        repo:    repo,
        auth:    auth,
        adminID: adminID,
    }
}

func (h *BotHandler) HandleUpdate(update tgbotapi.Update) {
    if update.Message == nil {
        return
    }

    msg := update.Message
    chatID := msg.Chat.ID
    
    // Аутентификация пользователя
    user, err := h.auth.AuthenticateTelegramUser(
        msg.From.ID,
        msg.From.UserName,
        msg.From.FirstName,
        msg.From.LastName,
    )
    
    if err != nil {
        h.sendMessage(chatID, "Ошибка авторизации. Попробуйте позже.")
        log.Printf("Auth error: %v", err)
        return
    }

    // Обработка команд
    switch {
    case msg.IsCommand():
        h.handleCommand(msg, user)
    case strings.HasPrefix(msg.Text, "/create"):
        h.handleCreateEvent(msg, user)
    default:
        h.handleTextMessage(msg, user)
    }
}

func (h *BotHandler) handleCommand(msg *tgbotapi.Message, user *models.User) {
    chatID := msg.Chat.ID
    
    switch msg.Command() {
    case "start":
        h.sendMessage(chatID, fmt.Sprintf(
            "Привет, %s!\nЯ бот для планирования мероприятий.\n\n"+
            "Доступные команды:\n"+
            "/events - показать все мероприятия\n"+
            "/create - создать новое мероприятие\n"+
            "/help - помощь\n\n"+
            "Для создания мероприятия напишите:\n"+
            "/create Название|Описание|Дата(YYYY-MM-DD)|Место",
            user.FirstName,
        ))
    
    case "help":
        h.sendMessage(chatID, 
            "*Помощь по командам:*\n\n"+
            "/start - начать работу\n"+
            "/events - список всех мероприятий\n"+
            "/myevents - мои мероприятия\n"+
            "/create - создать мероприятие\n"+
            "/admin - админ-панель (только для админов)\n"+
            "/help - эта справка\n\n"+
            "*Создание мероприятия:*\n"+
            "Напишите: /create Название|Описание|2024-12-31|Место проведения")
    
    case "events":
        h.handleShowEvents(chatID)
    
    case "myevents":
        h.handleMyEvents(chatID, user.TelegramID)
    
    case "admin":
        h.handleAdminPanel(chatID, user)
    
    default:
        h.sendMessage(chatID, "Неизвестная команда. Напишите /help для списка команд.")
    }
}

func (h *BotHandler) handleCreateEvent(msg *tgbotapi.Message, user *models.User) {
    chatID := msg.Chat.ID
    
    // Извлекаем данные из сообщения
    // Формат: /create Название|Описание|Дата|Место
    parts := strings.SplitN(msg.Text, " ", 2)
    if len(parts) < 2 {
        h.sendMessage(chatID, "Неверный формат. Используйте:\n"+
            "/create Название|Описание|2024-12-31|Место проведения")
        return
    }
    
    dataParts := strings.Split(parts[1], "|")
    if len(dataParts) != 4 {
        h.sendMessage(chatID, "Неверный формат. Нужно 4 части через |")
        return
    }
    
    // Парсим дату
    date, err := time.Parse("2006-01-02", strings.TrimSpace(dataParts[2]))
    if err != nil {
        h.sendMessage(chatID, "Неверный формат даты. Используйте YYYY-MM-DD")
        return
    }
    
    // Создаем мероприятие
    event := &models.Event{
        Title:       strings.TrimSpace(dataParts[0]),
        Description: strings.TrimSpace(dataParts[1]),
        Date:        date,
        Location:    strings.TrimSpace(dataParts[3]),
        CreatedBy:   user.TelegramID,
    }
    
    if err := h.repo.CreateEvent(event); err != nil {
        h.sendMessage(chatID, "Ошибка при создании мероприятия")
        log.Printf("Create event error: %v", err)
        return
    }
    
    h.sendMessage(chatID, fmt.Sprintf(
        "Мероприятие создано!\n\n"+
        "*Название:* %s\n"+
        "*Описание:* %s\n"+
        "*Дата:* %s\n"+
        "*Место:* %s",
        event.Title, event.Description, 
        event.Date.Format("02.01.2006"), event.Location,
    ))
}

func (h *BotHandler) handleShowEvents(chatID int64) {
    events, err := h.repo.GetAllEvents()
    if err != nil {
        h.sendMessage(chatID, "Ошибка при получении мероприятий")
        return
    }
    
    if len(events) == 0 {
        h.sendMessage(chatID, "Мероприятий пока нет")
        return
    }
    
    var response strings.Builder
    response.WriteString("*Все мероприятия:*\n\n")
    
    for _, event := range events {
        response.WriteString(fmt.Sprintf(
            "• *%s*\n  📍 %s\n  📅 %s\n  👤 Создатель: %d\n\n",
            event.Title, event.Location,
            event.Date.Format("02.01.2006 15:04"),
            event.CreatedBy,
        ))
    }
    
    h.sendMessage(chatID, response.String())
}

func (h *BotHandler) handleMyEvents(chatID, userID int64) {
    // В реальной реализации здесь был бы запрос по created_by
    // Для простоты показываем все
    events, err := h.repo.GetAllEvents()
    if err != nil {
        h.sendMessage(chatID, "❌ Ошибка при получении мероприятий")
        return
    }
    
    var myEvents []models.Event
    for _, event := range events {
        if event.CreatedBy == userID {
            myEvents = append(myEvents, event)
        }
    }
    
    if len(myEvents) == 0 {
        h.sendMessage(chatID, "У вас пока нет мероприятий")
        return
    }
    
    var response strings.Builder
    response.WriteString("*Ваши мероприятия:*\n\n")
    
    for _, event := range myEvents {
        response.WriteString(fmt.Sprintf(
            "• *%s*\n  📍 %s\n  📅 %s\n\n",
            event.Title, event.Location,
            event.Date.Format("02.01.2006 15:04"),
        ))
    }
    
    h.sendMessage(chatID, response.String())
}

func (h *BotHandler) handleAdminPanel(chatID int64, user *models.User) {
    // Проверка прав админа
    isAdmin, err := h.auth.IsAdmin(user.TelegramID)
    if err != nil || !isAdmin {
        h.sendMessage(chatID, "❌ У вас нет прав администратора")
        return
    }
    
    // Команды админа
    response := "*Админ-панель*\n\n" +
        "Доступные команды:\n" +
        "/admin_users - список пользователей\n" +
        "/admin_stats - статистика\n" +
        "/admin_makeadmin ID - назначить админом\n" +
        "/admin_delete_event ID - удалить мероприятие"
    
    h.sendMessage(chatID, response)
}

func (h *BotHandler) handleTextMessage(msg *tgbotapi.Message, user *models.User) {
    // Просто эхо-ответ для теста
    response := fmt.Sprintf("Вы написали: %s\n\nИспользуйте /help для списка команд", msg.Text)
    h.sendMessage(msg.Chat.ID, response)
}

func (h *BotHandler) sendMessage(chatID int64, text string) {
    msg := tgbotapi.NewMessage(chatID, text)
    msg.ParseMode = "Markdown"
    h.bot.Send(msg)
}
