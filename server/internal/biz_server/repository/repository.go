package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"server/internal/domain"
	"strings"
	"time"

	"github.com/jackc/pgx/v4"
)

var ErrUserNotFound = errors.New("user not found")

// описание структуры слоя репозитория
type BizRepository struct {
	DBRepo    *bizDBRepository
	CacheRepo *bizCacheRepository
}

// конструктор для слоя репозиторий
func NewBizRepository(dbRepo *bizDBRepository, cacheRepo *bizCacheRepository) (*BizRepository, error) {
	// Проверяем обязательные зависимости
	if dbRepo == nil {
		return nil, fmt.Errorf("dbRepo is required")
	}
	if cacheRepo == nil {
		return nil, fmt.Errorf("blackListRepo is required")
	}
	return &BizRepository{
		DBRepo:    dbRepo,
		CacheRepo: cacheRepo,
	}, nil
}

// метод для теста
func (r *BizRepository) Echo() string {
	return fmt.Sprintln("Hello from repo layer!")
}

// сохраняет или обновляет данные в таблице meesges
func (r *BizRepository) Save(ctx context.Context, message *domain.Message) error {
	// Определяем является ли сообщение командой
	isCommand := false
	commandName := ""
	if len(message.Text) > 0 && message.Text[0] == '/' {
		isCommand = true
		parts := strings.Split(message.Text, " ")
		commandName = parts[0]
	}

	// Сохраняем сообщение
	query := `
        INSERT INTO messages (
            telegram_message_id, telegram_chat_id, telegram_user_id,
            text, direction, status, is_command, command_name, created_at, updated_at
        ) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
        ON CONFLICT (telegram_chat_id, telegram_message_id) 
        DO UPDATE SET
            text = EXCLUDED.text,
            status = EXCLUDED.status,
            updated_at = EXCLUDED.updated_at
        RETURNING id
    `

	var id int64
	err := r.DBRepo.Pool.QueryRow(ctx, query,
		message.MessageID,
		message.ChatID,
		message.UserID,
		message.Text,
		message.Direction,
		message.Status,
		isCommand,
		commandName,
		message.CreatedAt,
		message.CreatedAt, // created_at и updated_at
	).Scan(&id)

	if err != nil {
		return fmt.Errorf("failed to save message: %w", err)
	}

	message.ID = id
	return nil
}

// SaveCallback сохраняет колбэк и связывает с сообщением
func (r *BizRepository) SaveCallback(ctx context.Context, callback *domain.CallbackLog) error {
	// Начинаем транзакцию
	tx, err := r.DBRepo.Pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	// Находим связанное сообщение (опционально)
	var messageID *int64
	err = tx.QueryRow(ctx, `
        SELECT id FROM messages 
        WHERE telegram_chat_id = $1 AND telegram_message_id = $2
    `, callback.ChatID, callback.MessageID).Scan(&messageID)

	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return fmt.Errorf("failed to find related message: %w", err)
	}

	// Сохраняем колбэк
	query := `
        INSERT INTO callback_logs (
            callback_id, telegram_user_id, telegram_chat_id, 
            telegram_message_id, callback_data, message_id, created_at
        ) VALUES ($1, $2, $3, $4, $5, $6, $7)
        ON CONFLICT (callback_id) DO NOTHING
        RETURNING id
    `

	var id int64
	err = tx.QueryRow(ctx, query,
		callback.CallbackID, callback.UserID, callback.ChatID,
		callback.MessageID, callback.Data, messageID,
		time.Now(),
	).Scan(&id)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			// Колбэк уже существует, это нормально для повторных обработок
			return nil
		}
		return fmt.Errorf("failed to save callback: %w", err)
	}

	// Коммитим транзакцию
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	callback.ID = id
	return nil
}

// метод для создания и сохранения пользователя в базу
func (r *BizRepository) CreateUser(ctx context.Context, user *domain.User) error {
	query := `
        INSERT INTO users (
            telegram_id, username, first_name, last_name, 
            is_active, created_at, last_seen_at
        ) VALUES ($1, $2, $3, $4, $5, $6, $7)
        RETURNING id
    `

	err := r.DBRepo.Pool.QueryRow(ctx, query,
		user.TelegramID,
		nullString(user.Username),
		user.FirstName,
		nullString(user.LastName),
		user.IsActive,
		user.CreatedAt,
		user.LastSeenAt,
	).Scan(&user.ID)

	if err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}

	return nil
}

// метод для обновления пользователя в базе (вдруг данные в телеграмме поменялись)
func (r *BizRepository) Update(ctx context.Context, user *domain.User) error {
	query := `
        UPDATE users SET
            username = $2,
            first_name = $3,
            last_name = $4,
            is_active = $5,
            last_seen_at = $6
        WHERE telegram_id = $1
        RETURNING id
    `

	err := r.DBRepo.Pool.QueryRow(ctx, query,
		user.TelegramID,
		nullString(user.Username),
		user.FirstName,
		nullString(user.LastName),
		user.IsActive,
		user.LastSeenAt,
	).Scan(&user.ID)

	if err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}

	return nil
}

// метод для поиска пользователя по ID из телеграмма
func (r *BizRepository) GetUserByTelegramID(ctx context.Context, telegramID int64) (*domain.User, error) {
	query := `
        SELECT id, telegram_id, username, first_name, last_name,
               is_active, created_at, last_seen_at
        FROM users
        WHERE telegram_id = $1
    `

	user := &domain.User{}
	var username, lastName sql.NullString

	err := r.DBRepo.Pool.QueryRow(ctx, query, telegramID).Scan(
		&user.ID,
		&user.TelegramID,
		&username,
		&user.FirstName,
		&lastName,
		&user.IsActive,
		&user.CreatedAt,
		&user.LastSeenAt,
	)

	if err != nil {
		// Проверяем, содержит ли ошибка "no rows in result set"
		if err.Error() == sql.ErrNoRows.Error() || strings.Contains(err.Error(), "no rows") {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	user.Username = username.String
	user.LastName = lastName.String

	return user, nil
}

// метод обновления времени последнего посещения пользователем по ID из телеграмм
func (r *BizRepository) UpdateLastSeen(ctx context.Context, telegramID int64) error {
	query := `UPDATE users SET last_seen_at = NOW() WHERE telegram_id = $1`

	_, err := r.DBRepo.Pool.Exec(ctx, query, telegramID)
	if err != nil {
		return fmt.Errorf("failed to update last_seen: %w", err)
	}

	return nil
}

// Вспомогательная функция
func nullString(s string) sql.NullString {
	if s == "" {
		return sql.NullString{Valid: false}
	}
	return sql.NullString{String: s, Valid: true}
}
