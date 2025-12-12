package database

import (
	"database/sql"
	"log"
	"os"

	"event-planner-bot/internal/models"

	_ "github.com/mattn/go-sqlite3"
)

// Хранилище для работы с БД
type Storage struct {
	db *sql.DB
}

// Создание хранилища
func NewStorage(dbPath string) (*Storage, error) {
	log.Println("Подключение к базе данных")

	// 1. Создание папки 'data', если ее нет
	if err := os.MkdirAll("data", 0755); err != nil {
		return nil, err
	}

	// 2. Открытие файла базы данных
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, err
	}

	// 3. Проверка подключения
	if err := db.Ping(); err != nil {
		return nil, err
	}

	// 4. Создание таблиц
	if err := createTables(db); err != nil {
		return nil, err
	}

	log.Println("База данных создана")
	return &Storage{db: db}, nil
}

// Создание таблиц
func createTables(db *sql.DB) error {
	log.Println("Создание таблиц")

	createUsersTable := `
    CREATE TABLE IF NOT EXISTS users (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        telegram_id INTEGER UNIQUE NOT NULL,
        username TEXT,
        first_name TEXT NOT NULL,
        last_name TEXT,
        is_admin BOOLEAN DEFAULT FALSE,
        created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
    );`

	createEventsTable := `
    CREATE TABLE IF NOT EXISTS events (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        title TEXT NOT NULL,
        description TEXT,
        date TIMESTAMP NOT NULL,
        location TEXT,
        created_by INTEGER NOT NULL,
        created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
        updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
    );`

	// Исправлено: правильные имена переменных
	if _, err := db.Exec(createUsersTable); err != nil {
		return err
	}

	if _, err := db.Exec(createEventsTable); err != nil {
		return err
	}

	log.Println("Таблицы созданы")
	return nil
}

// Создание пользователя (CreateUser вместо AddUser)
func (s *Storage) CreateUser(user *models.User) error {
	log.Printf("Создание пользователя: %s", user.Username)

	query := `INSERT INTO users (telegram_id, username, first_name, last_name, is_admin)
              VALUES (?, ?, ?, ?, ?)`

	_, err := s.db.Exec(query,
		user.TelegramID,
		user.Username,
		user.Name,    // Поле Name в User
		user.Surname, // Поле Surname в User
		user.IsAdmin)

	return err
}

// Получение пользователя по Telegram ID
func (s *Storage) GetUserByTelegramID(telegramID int64) (*models.User, error) {
	log.Printf("Поиск пользователя с ID: %d", telegramID)

	user := &models.User{}

	query := `
    SELECT id, telegram_id, username, first_name, last_name, is_admin, created_at
    FROM users
    WHERE telegram_id = ?`

	row := s.db.QueryRow(query, telegramID)

	err := row.Scan(
		&user.ID,
		&user.TelegramID,
		&user.Username,
		&user.Name,    // Маппинг на Name (в БД first_name)
		&user.Surname, // Маппинг на Surname (в БД last_name)
		&user.IsAdmin,
		&user.CreatedAt,
	)

	if err == sql.ErrNoRows {
		log.Println("Пользователь не найден")
		return nil, nil
	}

	return user, err
}

// Создание мероприятия
func (s *Storage) CreateEvent(event *models.Event) error {
	log.Printf("Создание мероприятия: %s", event.Title)

	query := `
    INSERT INTO events (title, description, date, location, created_by)
    VALUES (?, ?, ?, ?, ?)`

	_, err := s.db.Exec(query,
		event.Title,
		event.Description,
		event.EventDate, // Внимание: поле EventDate, а не Date!
		event.Location,
		event.CreatedBy)

	return err
}

// Получение всех мероприятий
func (s *Storage) GetAllEvents() ([]models.Event, error) {
	log.Println("Получение всех мероприятий")

	query := `
    SELECT id, title, description, date, location, created_by, created_at, updated_at
    FROM events
    ORDER BY date`

	rows, err := s.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var events []models.Event

	for rows.Next() {
		var event models.Event
		err := rows.Scan(
			&event.ID,
			&event.Title,
			&event.Description,
			&event.EventDate, // Внимание: поле EventDate!
			&event.Location,
			&event.CreatedBy,
			&event.CreatedAt,
			&event.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		events = append(events, event)
	}

	log.Printf("Найдено %d мероприятий", len(events))
	return events, nil
}

// Получение мероприятия по ID
func (s *Storage) GetEventByID(id int64) (*models.Event, error) {
	log.Printf("Поиск мероприятия с ID: %d", id)

	event := &models.Event{}

	query := `
    SELECT id, title, description, date, location, created_by, created_at, updated_at
    FROM events
    WHERE id = ?`

	row := s.db.QueryRow(query, id)

	err := row.Scan(
		&event.ID,
		&event.Title,
		&event.Description,
		&event.EventDate, // Внимание: поле EventDate!
		&event.Location,
		&event.CreatedBy,
		&event.CreatedAt,
		&event.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		log.Println("Мероприятие не найдено")
		return nil, nil
	}

	return event, err
}

// Обновление мероприятия
func (s *Storage) UpdateEvent(event *models.Event) error {
	log.Printf("Обновление мероприятия ID: %d", event.ID)

	query := `
    UPDATE events
    SET title = ?, description = ?, date = ?, location = ?, updated_at = CURRENT_TIMESTAMP
    WHERE id = ?`

	_, err := s.db.Exec(query,
		event.Title,
		event.Description,
		event.EventDate, // Внимание: поле EventDate!
		event.Location,
		event.ID)

	return err
}

// Удаление мероприятия
func (s *Storage) DeleteEvent(id int64) error {
	log.Printf("Удаление мероприятия ID: %d", id)

	query := `DELETE FROM events WHERE id = ?`
	_, err := s.db.Exec(query, id)
	return err
}

// Закрытие соединения
func (s *Storage) Close() {
	log.Println("Закрытие соединения с БД")
	s.db.Close()
}
