package handlersgrpc

import (
	"context"
	"fmt"
	"server/internal/biz_server/service"
	"server/internal/domain"
	"time"

	pb "global_models/grpc/bot"
)

type GRPCHandlerInterface interface {
	// ProcessMessage - обработка входящего сообщения
	ProcessMessage(ctx context.Context, msg *pb.Message) (*pb.UpdateResponse, error)

	// ProcessCallback - обработка callback от inline клавиатуры
	ProcessCallback(ctx context.Context, callback *pb.CallbackQuery) (*pb.UpdateResponse, error)
}

// На этом слое остается только транспортная логика (преобразование данных и управление запросом/ответом)
type BizGRPCHandler struct {
	Service service.InMessageServiceInterface
}

func NewBizGRPCHandler(grpcService service.InMessageServiceInterface) GRPCHandlerInterface {
	return &BizGRPCHandler{
		Service: grpcService,
	}
}

// ProcessMessage - обработка входящего сообщения
func (b *BizGRPCHandler) ProcessMessage(ctx context.Context, msg *pb.Message) (*pb.UpdateResponse, error) {
	// Приводим входящее сообщение к domain типу
	incomingMsg := &domain.Message{
		MessageID: msg.MessageId,
		ChatID:    msg.ChatId,
		UserID:    msg.UserId,
		Text:      msg.Text,
		Direction: "incoming",
		Timestamp: time.Unix(msg.Date, 0),
	}

	// Сохранаяем сообщение в БД через сервисный слой
	err := b.Service.CheckAndSaveMsg(incomingMsg)
	if err != nil {
		return nil, fmt.Errorf("failed to save incoming message: %w", err)
	}

	// генерируем ответ в сервисном слое
	replyText := b.Service.GenerateReply(incomingMsg.Text, msg.From)

	// создаём исходящее сообщение на базе domain типа, чтобы его сохранить
	outgoingMsg := &domain.Message{
		ChatID:    msg.ChatId,
		UserID:    msg.UserId,
		Text:      replyText,
		Direction: "outgoing",
		Timestamp: time.Now(),
	}

	// Сохранаяем сообщение в БД через сервисный слой
	err = b.Service.CheckAndSaveMsg(outgoingMsg)
	if err != nil {
		return nil, fmt.Errorf("failed to save outgoing message: %w", err)
	}

	// Формируем ответ для бота

	response := &pb.UpdateResponse{
		Success: true,
		Messages: []*pb.OutgoingMessage{
			{
				ChatId: msg.ChatId,
				Text:   replyText,
				// Можно добавить клавиатуру если нужно
				ReplyMarkup: b.Service.CreateTestKeyboard(),
			},
		},
	}

	return response, nil
}

// ProcessCallback - обработка callback от inline клавиатуры
func (b *BizGRPCHandler) ProcessCallback(ctx context.Context, callback *pb.CallbackQuery) (*pb.UpdateResponse, error) {
	// Приводим входящий callback к domain типу
	callBackLog := &domain.CallbackLog{
		CallbackID: callback.Id,
		UserID:     callback.UserId,
		ChatID:     callback.ChatId,
		MessageID:  callback.MessageId,
		Data:       callback.Data,
		Timestamp:  time.Now(),
	}

	// проверяем и сохраняем данные в БД
	err := b.Service.CheckAndSaveCallBack(callBackLog)
	if err != nil {
		return nil, fmt.Errorf("failed to save callback: %w", err)
	}

	// Анализируем данные callback и формируем ответ
	response := &pb.UpdateResponse{
		Success: true,
	}

	switch callback.Data {
	case "help":
		// Отправляем новое сообщение с помощью
		response.Messages = append(response.Messages, &pb.OutgoingMessage{
			ChatId: callback.ChatId,
			Text:   "Я бот-помощник. Доступные команды:\n/help - помощь\n/start - начало",
		})

	case "more":
		// Редактируем существующее сообщение (меняем текст)
		// Для редактирования нужно добавить поле в protobuf, пока просто шлем новое
		response.Messages = append(response.Messages, &pb.OutgoingMessage{
			ChatId: callback.ChatId,
			Text:   "Дополнительная информация...",
		})

	default:
		// Ответ на неизвестную команду
		response.Messages = append(response.Messages, &pb.OutgoingMessage{
			ChatId: callback.ChatId,
			Text:   fmt.Sprintf("Неизвестная команда: %s", callback.Data),
		})
	}
	return response, nil
}
