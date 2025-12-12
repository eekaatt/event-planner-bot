package config

import "os"

// getEnv получает переменную окружения или значение по умолчанию
func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

type Config struct {
	TelegramToken string
	DBPath        string
	AdminID       int64
	Debug         bool
}

func LoadConfig() (*Config, error) {
	return &Config{
		TelegramToken: getEnv("TELEGRAM_BOT_TOKEN", "8250977349:AAHPQwyMLuhH5obsa8r59xLoiuxjOLbI8gw"),
		DBPath:        getEnv("DB_PATH", "./data/events.db"),
		AdminID:       2025081326,
		Debug:         getEnv("DEBUG", "true") == "true",
	}, nil
}
