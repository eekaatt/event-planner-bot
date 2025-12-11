package models

import ("time")

type User struct {
	ID int64 `json:"id"`  // id пользователя
	TelegramID int64 `json:"telegram_id"` // телеграмм id пользователя
	Username string `json:"username"` // имя пользователя в телеграмме
	Name string `json:"name"`  // имя пользователя
	Surname string `json:"surname"`  // фамилия пользователя
	IsAdmin bool `json:"is_admin"`  // является ли админом
	CreatedAt time.Time `json:"created_at"`  // время первого захода в бот
}

// Роли пользователей: либо админ, либо пользователь
type UserRole string

const (
	RoleUser UserRole = "user"
	RoleAdmin UserRole = "admin"
)
