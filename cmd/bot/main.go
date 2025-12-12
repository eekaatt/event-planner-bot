package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"event-planner-bot/config"
	"event-planner-bot/internal/auth"
	"event-planner-bot/internal/bot"
	"event-planner-bot/internal/database"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func main() {
	log.Println("Запуск EventBot...")

	// Загружаем конфигурацию
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Ошибка загрузки конфигурации: %v", err)
	}

	if cfg.TelegramToken == "" {
		log.Fatal("TELEGRAM_BOT_TOKEN не установлен")
	}

	// Подключаемся к базе данных
	repo, err := database.NewStorage(cfg.DBPath)
	if err != nil {
		log.Fatalf("Ошибка подключения к БД: %v", err)
	}
	defer repo.Close()

	log.Printf("База данных: %s", cfg.DBPath)

	// Создаем сервис аутентификации
	authService := auth.NewAuthService(repo)

	// Создаем бота
	botAPI, err := tgbotapi.NewBotAPI(cfg.TelegramToken)
	if err != nil {
		log.Fatalf("Ошибка создания бота: %v", err)
	}

	botAPI.Debug = cfg.Debug
	log.Printf("Авторизован как %s", botAPI.Self.UserName)

	// Создаем обработчик
	botHandler := bot.NewBotHandler(botAPI, repo, authService, cfg.AdminID)

	// Настраиваем обновления
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates := botAPI.GetUpdatesChan(u)

	// Обработка сигналов для graceful shutdown
	stopChan := make(chan os.Signal, 1)
	signal.Notify(stopChan, os.Interrupt, syscall.SIGTERM)

	log.Println("Бот запущен. Ожидаем сообщения...")

	// Главный цикл обработки сообщений
	for {
		select {
		case update := <-updates:
			go botHandler.HandleUpdate(update)

		case <-stopChan:
			log.Println("Остановка бота...")
			botAPI.StopReceivingUpdates()
			return
		}
	}
}
