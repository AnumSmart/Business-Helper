package domain

import "time"

// Внутренние структуры для БД (не protobuf)
type Message struct {
	ID          int64     `db:"id"`
	MessageID   int64     `db:"telegram_message_id"`
	ChatID      int64     `db:"telegram_chat_id"`
	UserID      int64     `db:"telegram_user_id"`
	Text        string    `db:"text"`
	Direction   string    `db:"direction"` // "incoming" или "outgoing"
	Status      string    `db:"status"`    // "sent", "delivered", "read", "failed", "pending"
	IsCommand   bool      `db:"is_command"`
	CommandName string    `db:"command_name"`
	CreatedAt   time.Time `db:"created_at"`
	UpdatedAt   time.Time `db:"updated_at"`
	TimeStamp   time.Time
}

type CallbackLog struct {
	ID         int64     // внутренний ID в БД (не из proto!)
	CallbackID string    // = Id из proto
	UserID     int64     // = UserId из proto
	ChatID     int64     // = ChatId из proto
	MessageID  int64     // = MessageId из proto
	Data       string    // = Data из proto
	Timestamp  time.Time // добавляем время получения
}

// IncomingMessage - входящее сообщение от другого сервера
type IncomingMessage struct {
	ChatID      int64        // = ChatId из proto
	Text        string       // = Text из proto
	ReplyMarkup *ReplyMarkup // конвертированная клавиатура
	ReceivedAt  time.Time    // добавляем время получения
}

// ReplyMarkup - клавиатура
type ReplyMarkup struct {
	InlineKeyboard  [][]InlineButton
	Keyboard        [][]Button
	ResizeKeyboard  bool
	OneTimeKeyboard bool
}

// InlineButton - кнопка под сообщением
type InlineButton struct {
	Text         string
	CallbackData string
	URL          string
}

// Button - обычная кнопка
type Button struct {
	Text string
}

// MessageResponse - ответ от сервиса
type MessageResponse struct {
	Success bool
	Error   string
	// Можно добавить другие поля, если нужно вернуть что-то еще
}

// User - внутренняя модель пользователя Telegram
// Полностью соответствует таблице users из миграции
type User struct {
	ID         int64     // Внутренний ID в БД (BIGSERIAL)
	TelegramID int64     // Telegram ID (уникальный)
	Username   string    // Username (может быть пустым, без @)
	FirstName  string    // Имя
	LastName   string    // Фамилия (может быть пустой)
	IsActive   bool      // Активен ли пользователь
	CreatedAt  time.Time // Когда впервые появился
	LastSeenAt time.Time // Последняя активность
}
