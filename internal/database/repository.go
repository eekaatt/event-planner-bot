package database

import {
	"database/sql"  // для работы с базой данных
	"os"  // для работы с файлами
	"log"  // для вывода сообщений
	"time" // для работы со временем
	
	// Модели (User и Event)
    "github.com/eekaatt/event_planner_bot-go/internal/models"
    
    // Драйвер для SQLite (база данных в файле)
        _ "github.com/mattn/go-sqlite3"
}

// Хранилище для работы с БД
type Storage struct {
	db *sql.DB  // соединение с базой данных
}

func NewStorage(dbPath string) (*Storage, error) {
	log Println("Подключение к базе данных")
	
	// 1. Создание папки 'data', если ее нет
	if err := os.MkdirAll("data", 0755); err != nil {
		return nil, err
	}
	
	// 2. Открытие файла базы данных
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, err
	}
	
	// 3. Проверка, что подключение работает
	if err := db.Ping(); err != nil {
		return nil, err
	}
	
	// 4. Создание таблиц в БД
	if err := createTables(db); err != nil {
		return nil, err
	}
	log.Println("База данных создана")
	return &Storage{db: db}, nil
}

// Создание таблицы в БД
func createTables(db *sql.DB) error {
	log.Println("Создание таблицы")
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
        
        if _, err := db.Exec(createUserTable); err != nil {
        	return err
        }
        
        if _, err := db.Exec(createEventsTable); err != nil {
        	return err
        }
        
        log.Println("Таблицы созданы")
        return nil
}

// Добавление пользователей
func (s *Storage) AddUser(user *models.User) error {
	log.Println("Добавление пользователя: ", user.Username)
	
	query := `INSERT INTO users (telegram_id, username, name, surname, is_admin)
    VALUES (?, ?, ?, ?, ?)`
    
    _, err := s.db.Exec(query, 
    user.TelegramID, user.Username, user.Name, user.Surname, user.IsAdmin)
    return err
}

// Найти пользователя по его ID телеграмма
func (s *Storage) FindUserByTelegramID(telegramID int64) (*models.User, error){
	log.Printf("Поиск пользователя с ID: ", telegramID)
    
    user := &models.User{}
    
    query := `
    SELECT id, telegram_id, username, first_name, last_name, is_admin, created_at
    FROM users
    WHERE telegram_id = ?`
    
    // Выполняем запрос и получаем одну строку
    row := s.db.QueryRow(query, telegramID)
    
    // Копируем данные из базы в нашу структуру
    err := row.Scan(
        &user.ID,
        &user.TelegramID,
        &user.Username,
        &user.FirstName,
        &user.LastName,
        &user.IsAdmin,
        &user.CreatedAt,
    )
    
    // Если пользователь не найден - это не ошибка
    if err == sql.ErrNoRows {
        log.Println("Пользователь не найден")
        return nil, nil
    }
    
    return user, err
}

// CRUD для мероприятий
// Добавление нового мероприятия
func (s *Storage) AddEvent(event *models.Event) error {
    log.Printf("Добавление мероприятия: ", event.Title)
    
    query := `
    INSERT INTO events (title, description, date, location, created_by)
    VALUES (?, ?, ?, ?, ?)`
    
    _, err := s.db.Exec(query,
        event.Title, event.Description, event.Date, event.Location, event.CreatedBy)
    
    return err
}

// Получение всех мероприятий
func (s *Storage) GetAllEvents() ([]models.Event, error) {
    log.Println("Получение всех мероприятий")
    
    query := `
    SELECT id, title, description, date, location, created_by, created_at, updated_at
    FROM events
    ORDER BY date`
    
    // Выполняем запрос (может вернуть много строк)
    rows, err := s.db.Query(query)
    if err != nil {
        return nil, err
    }
    
    // Закрытие соединения, когда закончим
    defer rows.Close()
    
    // Создаем пустой список мероприятий
    events := []models.Event{}
    
    // Перебираем все строки результата
    for rows.Next() {
        var event models.Event
        
        // Копируем данные из строки в структуру
        err := rows.Scan(
            &event.ID,
            &event.Title,
            &event.Description,
            &event.Date,
            &event.Location,
            &event.CreatedBy,
            &event.CreatedAt,
            &event.UpdatedAt,
        )
        
        if err != nil {
            return nil, err
        }
        
        // Добавление мероприятия в список
        events = append(events, event)
    }
    
    log.Printf("Найдено %d мероприятий", len(events))
    return events, nil
}

// Поиск мероприятия по ID
func (s *Storage) FindEventByID(id int64) (*models.Event, error) {
    log.Printf("Поиск мероприятие с ID: %d", id)
    
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
        &event.Date,
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
    log.Printf("Обновление мероприятия с ID: %d", event.ID)
    
    query := `
    UPDATE events
    SET title = ?, description = ?, date = ?, location = ?, updated_at = CURRENT_TIMESTAMP
    WHERE id = ?`
    
    _, err := s.db.Exec(query,
        event.Title,
        event.Description,
        event.Date,
        event.Location,
        event.ID)
    
    return err
}

// Удаление мероприятия
func (s *Storage) DeleteEvent(id int64) error {
    log.Printf("Удаление мероприятие с ID: %d", id)
    
    query := `DELETE FROM events WHERE id = ?`
    _, err := s.db.Exec(query, id)
    
    return err
}

// Закрытие соединения с БД
func (s *Storage) Close() {
    log.Println("Закрытие соединения с БД")
    s.db.Close()
}
