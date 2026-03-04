package converter

import (
	"bot/internal/domain"

	tele "gopkg.in/telebot.v4"
)

// конвертируем информацию из контеста сообщения от бота при poling режиме, приводим к доменному типу
func ConvertToUpdate(ctx tele.Context) (*domain.TelegramUpdate, error) {
	update := &domain.TelegramUpdate{
		UpdateID: int64(ctx.Update().ID), // получаем ID обновления
	}

	// Определяем тип обновления и заполняем соответствующие поля
	switch {
	case ctx.Message() != nil: // Это обычное сообщение
		fillMessage(update, ctx)
	case ctx.Callback() != nil: // Это callback от inline кнопки
		fillCallback(update, ctx)
	}

	return update, nil
}

// вспомогательная функция заполнения при конвертации, если это сообщение
func fillMessage(update *domain.TelegramUpdate, ctx tele.Context) {
	msg := ctx.Message()

	update.Message = &struct {
		MessageID int64 `json:"message_id"`
		From      struct {
			ID       int64  `json:"id"`
			Username string `json:"username"`
		} `json:"from"`
		Chat struct {
			ID int64 `json:"id"`
		} `json:"chat"`
		Date int64  `json:"date"`
		Text string `json:"text"`
	}{
		MessageID: int64(msg.ID),
		Date:      int64(msg.Unixtime),
		Text:      msg.Text,
	}

	// Заполняем информацию об отправителе
	if msg.Sender != nil {
		update.Message.From.ID = msg.Sender.ID
		update.Message.From.Username = msg.Sender.Username
	}

	// Заполняем информацию о чате
	if msg.Chat != nil {
		update.Message.Chat.ID = msg.Chat.ID
	}
}

// вспомогательная функция заполнения при конвертации, если это сколбэк от inline клавиатуры
func fillCallback(update *domain.TelegramUpdate, ctx tele.Context) {
	callback := ctx.Callback()

	update.CallbackQuery = &struct {
		ID   string `json:"id"`
		From struct {
			ID int64 `json:"id"`
		} `json:"from"`
		Message struct {
			MessageID int64 `json:"message_id"`
			Chat      struct {
				ID int64 `json:"id"`
			} `json:"chat"`
		} `json:"message"`
		Data string `json:"data"`
	}{
		ID:   callback.ID,
		Data: callback.Data,
	}

	// Заполняем информацию об отправителе
	if callback.Sender != nil {
		update.CallbackQuery.From.ID = callback.Sender.ID
	}

	// Заполняем информацию о сообщении и чате
	if callback.Message != nil {
		update.CallbackQuery.Message.MessageID = int64(callback.Message.ID)
		if callback.Message.Chat != nil {
			update.CallbackQuery.Message.Chat.ID = callback.Message.Chat.ID
		}
	}
}
