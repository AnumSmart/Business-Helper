package domain

// User представляет информацию о пользователе Telegram
type User struct {
	ID       int64  `json:"id"`
	Username string `json:"username"`
	// Добавьте другие поля при необходимости:
	FirstName string `json:"first_name,omitempty"`
	LastName  string `json:"last_name,omitempty"`
	IsBot     bool   `json:"is_bot,omitempty"`
}

// Chat представляет информацию о чате Telegram
type Chat struct {
	ID int64 `json:"id"`
	// Добавьте другие поля при необходимости:
	// Type string `json:"type,omitempty"`
	// Title string `json:"title,omitempty"`
}

// Message представляет сообщение Telegram
type Message struct {
	MessageID int64  `json:"message_id"`
	From      User   `json:"from"`
	Chat      Chat   `json:"chat"`
	Date      int64  `json:"date"`
	Text      string `json:"text,omitempty"`
}

// CallbackQuery представляет callback запрос от inline клавиатуры
type CallbackQuery struct {
	ID      string  `json:"id"`
	From    User    `json:"from"`
	Message Message `json:"message"`
	Data    string  `json:"data"`
}

// TelegramUpdate представляет структуру входящего обновления от Telegram API
type TelegramUpdate struct {
	UpdateID      int64          `json:"update_id"`
	Message       *Message       `json:"message,omitempty"`
	CallbackQuery *CallbackQuery `json:"callback_query,omitempty"`
}
