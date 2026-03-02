package domain

// TelegramUpdate представляет структуру входящего обновления от Telegram API
// Теги json используются для маппинга полей из JSON в структуру
type TelegramUpdate struct {
	UpdateID int64 `json:"update_id"` // Уникальный ID обновления

	// Message - указатель, т.к. может отсутствовать (если это callback)
	Message *struct {
		MessageID int64 `json:"message_id"` // ID сообщения в чате
		From      struct {
			ID       int64  `json:"id"`       // ID отправителя
			Username string `json:"username"` // Username (без @)
		} `json:"from"`
		Chat struct {
			ID int64 `json:"id"` // ID чата
		} `json:"chat"`
		Date int64  `json:"date"` // Unix timestamp
		Text string `json:"text"` // Текст сообщения
	} `json:"message"`

	// CallbackQuery - указатель, т.к. может отсутствовать (если это сообщение)
	CallbackQuery *struct {
		ID   string `json:"id"` // Уникальный ID callback
		From struct {
			ID int64 `json:"id"` // ID пользователя, нажавшего кнопку
		} `json:"from"`
		Message struct {
			MessageID int64 `json:"message_id"` // ID сообщения с клавиатурой
			Chat      struct {
				ID int64 `json:"id"` // ID чата
			} `json:"chat"`
		} `json:"message"`
		Data string `json:"data"` // Данные из callback_data
	} `json:"callback_query"`
}
