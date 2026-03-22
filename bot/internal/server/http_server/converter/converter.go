package converter

import (
	"bot/internal/domain"
	pb "global_models/grpc/bot"
)

// ConvertToGRPCUpdate конвертирует Telegram структуру в protobuf структуру
// Это "адаптер" между внешним API (Telegram) и внутренним (gRPC)
func ConvertToGRPCUpdate(update *domain.TelegramUpdate) *pb.UpdateRequest {
	// Создаем базовый запрос с update_id
	req := &pb.UpdateRequest{
		UpdateId: update.UpdateID,
	}

	// Если есть сообщение - заполняем структуру Message
	if update.Message != nil {
		req.Message = &pb.Message{
			MessageId: update.Message.MessageID,
			ChatId:    update.Message.Chat.ID,
			UserId:    update.Message.From.ID,
			Text:      update.Message.Text,
			Date:      update.Message.Date,
			From: &pb.User{
				Id:       update.Message.From.ID,
				Username: update.Message.From.Username,
			},
			Chat: &pb.Chat{
				Id:   update.Message.Chat.ID,
				Type: "private", // Упрощение: в реальном проекте нужно определять тип
			},
		}
	}

	// Если есть callback query - заполняем структуру CallbackQuery
	if update.CallbackQuery != nil {
		req.CallbackQuery = &pb.CallbackQuery{
			Id:        update.CallbackQuery.ID,
			UserId:    update.CallbackQuery.From.ID,
			MessageId: update.CallbackQuery.Message.MessageID,
			ChatId:    update.CallbackQuery.Message.Chat.ID,
			Data:      update.CallbackQuery.Data,
			From: &pb.User{
				Id:       update.CallbackQuery.From.ID,
				Username: update.CallbackQuery.From.Username,
			},
		}
	}

	return req
}
