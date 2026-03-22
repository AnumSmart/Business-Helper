package converter

import (
	"bot/internal/domain"
	"fmt"

	tele "gopkg.in/telebot.v4"
)

// ConvertToUpdate конвертирует контекст телебота в доменную структуру TelegramUpdate
func ConvertToUpdate(ctx tele.Context) (*domain.TelegramUpdate, error) {
	update := &domain.TelegramUpdate{
		UpdateID: int64(ctx.Update().ID),
	}

	switch {
	case ctx.Callback() != nil:
		fillCallback(update, ctx)
	case ctx.Message() != nil:
		fillMessage(update, ctx)
	default:
		return nil, fmt.Errorf("неподдерживаемый тип обновления")
	}

	return update, nil
}

// fillMessage заполняет структуру сообщения
func fillMessage(update *domain.TelegramUpdate, ctx tele.Context) {
	msg := ctx.Message()
	if msg == nil {
		return
	}

	update.Message = &domain.Message{
		MessageID: int64(msg.ID),
		Date:      int64(msg.Unixtime),
		Text:      msg.Text,
	}

	// Заполняем информацию об отправителе
	if msg.Sender != nil {
		update.Message.From = domain.User{
			ID:       msg.Sender.ID,
			Username: msg.Sender.Username,
		}
	}

	// Заполняем информацию о чате
	if msg.Chat != nil {
		update.Message.Chat = domain.Chat{
			ID: msg.Chat.ID,
		}
	}
}

// fillCallback заполняет структуру callback запроса
func fillCallback(update *domain.TelegramUpdate, ctx tele.Context) {
	callback := ctx.Callback()
	if callback == nil {
		return
	}

	update.CallbackQuery = &domain.CallbackQuery{
		ID:   callback.ID,
		Data: callback.Data,
	}

	// Заполняем информацию об отправителе
	if callback.Sender != nil {
		update.CallbackQuery.From = domain.User{
			ID:       callback.Sender.ID,
			Username: callback.Sender.Username,
		}
	}

	// Заполняем информацию о сообщении
	if callback.Message != nil {
		update.CallbackQuery.Message = domain.Message{
			MessageID: int64(callback.Message.ID),
		}

		if callback.Message.Chat != nil {
			update.CallbackQuery.Message.Chat = domain.Chat{
				ID: callback.Message.Chat.ID,
			}
		}
	}
}
