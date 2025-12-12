package auth

import (
	"log"

	"event-planner-bot/internal/database"
	"event-planner-bot/internal/models"
)

type AuthService struct {
	repo *database.Storage // Вместо *database.Repository
}

func NewAuthService(repo *database.Storage) *AuthService {
	return &AuthService{repo: repo}
}

// Регистрация/логин пользователя Telegram
func (a *AuthService) AuthenticateTelegramUser(telegramID int64, username, firstName, lastName string) (*models.User, error) {
	// Проверяем существующего пользователя
	user, err := a.repo.GetUserByTelegramID(telegramID)
	if err != nil {
		return nil, err
	}

	// Если пользователь не найден - создаем нового
	if user == nil {
		user = &models.User{
			TelegramID: telegramID,
			Username:   username,
			Name:       firstName,
			Surname:    lastName,
			IsAdmin:    false, // По умолчанию не админ
		}

		if err := a.repo.CreateUser(user); err != nil {
			return nil, err
		}

		log.Printf("Создан новый пользователь: %s (ID: %d)", username, telegramID)
	} else {
		// Обновляем информацию, если нужно
		log.Printf("Пользователь авторизован: %s (ID: %d)", username, telegramID)
	}

	return user, nil
}

// Проверка админских прав
func (a *AuthService) IsAdmin(telegramID int64) (bool, error) {
	user, err := a.repo.GetUserByTelegramID(telegramID)
	if err != nil || user == nil {
		return false, err
	}
	return user.IsAdmin, nil
}

// Назначение админа (только для существующих админов)
func (a *AuthService) MakeAdmin(telegramID int64) error {
	// В реальном приложении здесь была бы сложная логика
	// Для простоты: если пользователь существует, делаем его админом
	user, err := a.repo.GetUserByTelegramID(telegramID)
	if err != nil || user == nil {
		return err
	}

	// В реальности здесь должен быть SQL UPDATE
	user.IsAdmin = true
	return nil
}
