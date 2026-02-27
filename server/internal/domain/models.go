package domain

import "time"

// Внутренние структуры для БД (не protobuf)
type Message struct {
	ID        int64
	MessageID int64
	ChatID    int64
	UserID    int64
	Text      string
	Direction string // "incoming" или "outgoing"
	Status    string // "sent", "pending", "failed"
	Timestamp time.Time
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
// Не зависит от protobuf, используется во всем сервисном слое
type User struct {
	ID        int64  // Уникальный ID пользователя в Telegram
	FirstName string // Имя
	LastName  string // Фамилия (может быть пустой)
	Username  string // Username (может быть пустым, без @)
}
