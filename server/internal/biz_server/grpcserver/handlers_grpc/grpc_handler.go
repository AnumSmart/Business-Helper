package handlersgrpc

import (
	"context"
	"fmt"
	"server/internal/biz_server/grpcserver/converter"
	"server/internal/biz_server/service"
	"server/internal/domain"
	"server/internal/interfaces"
	"time"

	pb "global_models/grpc/bot"
)

// На этом слое остается только транспортная логика (преобразование данных и управление запросом/ответом)
type BizGRPCHandler struct {
	Service service.ServiceForGRPCHandler
}

func NewBizGRPCHandler(grpcService service.ServiceForGRPCHandler) interfaces.GRPCHandlerInterface {
	return &BizGRPCHandler{
		Service: grpcService,
	}
}

// ProcessMessage - обработка входящего сообщения
func (b *BizGRPCHandler) ProcessMessage(ctx context.Context, msg *pb.Message) (*pb.UpdateResponse, error) {
	// Приводим входящее сообщение к domain типу, используем конвертер
	incomingMsg := converter.ToDomainMessage(msg)

	// Сохранаяем сообщение в БД через сервисный слой
	err := b.Service.CheckAndSaveMsg(incomingMsg)
	if err != nil {
		return nil, fmt.Errorf("failed to save incoming message: %w", err)
	}

	// конвертируем пользователя из proto сообщения в доменную модель
	user := converter.ToDomainUser(msg.From)

	// генерируем ответ в сервисном слое
	replyText := b.Service.GenerateReply(incomingMsg.Text, user)

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
			},
		},
	}

	return response, nil
}

// ProcessCallback - обработка callback от inline клавиатуры
func (b *BizGRPCHandler) ProcessCallback(ctx context.Context, callback *pb.CallbackQuery) (*pb.UpdateResponse, error) {
	// Приводим входящий callback к domain типуб используем конвертер
	callBackLog := converter.ToCallbackLog(callback)

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
			Text:   "Я бот-помощник. Доступные команды:\n/help - помощь\n",
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

// Принятое сообщение от grpc клиента обрабатывается в слое хэндлера и передаётся в слой сервиса
func (b *BizGRPCHandler) ProcessIncomingMsg(ctx context.Context, req *pb.SendMessageRequest) (*pb.SendMessageResponse, error) {
	// 1. Валидация протокола
	if req == nil {
		return nil, fmt.Errorf("request is nil!")
	}

	// 2. Проверка обязательных полей на уровне протокола
	if req.ChatId == 0 {
		return nil, fmt.Errorf("Chat ID must not be 0!")
	}

	// 3. КОНВЕРТАЦИЯ: из protobuf во внутреннюю модель
	incomingMsg := converter.ToIncomingMessage(req)

	// 4. Вызов сервисного слоя с внутренней моделью
	result, err := b.Service.AnswerIncomingMsg(ctx, incomingMsg)
	if err != nil {
		return nil, err
	}

	// 5. КОНВЕРТАЦИЯ: из внутренней модели обратно в protobuf
	return converter.ToProtoResponse(result), nil
}
