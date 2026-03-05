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

	isCommand := false
	commandName := ""
	if len(message.Text) > 0 && message.Text[0] == '/' {
		isCommand = true
		// Простое извлечение команды (можно улучшить)
		parts := strings.Split(message.Text, " ")
		commandName = parts[0]
	}

	var id int64
	err := r.DBRepo.Pool.QueryRow(ctx, query,
		message.MessageID, message.ChatID, message.UserID,
		message.Text, message.Direction, message.Status,
		isCommand, commandName,
		message.Timestamp, message.Timestamp,
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
	var messageID sql.NullInt64
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
