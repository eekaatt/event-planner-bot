package config

import (
    "os"
    "strconv"

    "github.com/joho/godotenv"
)

type Config struct {
    TelegramToken string
    DBPath        string
    AdminID       int64
    ServerPort    string
    Debug         bool
}

func LoadConfig() (*Config, error) {
    // Загружаем .env файл
    godotenv.Load()
    
    adminID, _ := strconv.ParseInt(os.Getenv("ADMIN_TELEGRAM_ID"), 10, 64)
    
    return &Config{
        TelegramToken: os.Getenv("TELEGRAM_BOT_TOKEN"),
        DBPath:        os.Getenv("DB_PATH"),
        AdminID:       adminID,
        ServerPort:    os.Getenv("SERVER_PORT"),
        Debug:         os.Getenv("DEBUG") == "true",
    }, nil
}
