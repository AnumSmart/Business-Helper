package handlersgrpc

import (
	"context"
	"fmt"
	"log"
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

	// Добавим подробное логирование для отладки
	log.Printf("🔍 ProcessMessage получил:, UserID=%d, ChatID=%d, MessageID=%d, Text=%s",
		msg.UserId, msg.ChatId, msg.MessageId, msg.Text)

	// при вызове этого метода необходимо сохранить нового пользователя в базе или обновить некоторые его поля
	user, err := b.Service.RegisterOrUpdate(ctx, incomingMsg.MessageID, incomingMsg.UserFirstName, incomingMsg.UserLastName, incomingMsg.UserNickName)
	if err != nil {
		log.Printf("failed to save/update user: %v", err)
	}

	// Сохранаяем сообщение в БД через сервисный слой
	err = b.Service.CheckAndSaveMsg(ctx, incomingMsg)
	if err != nil {
		log.Printf("failed to save incoming message: %v", err)
		//return nil, fmt.Errorf("failed to save incoming message: %w", err)
	}

	// конвертируем пользователя из proto сообщения в доменную модель
	user = converter.ToDomainUser(msg.From)

	// генерируем ответ на текстовое сообщение в сервисном слое
	replyText := b.Service.GenerateReply(incomingMsg.Text, user)

	// создаём исходящее сообщение на базе domain типа, чтобы его сохранить
	outgoingMsg := &domain.Message{
		ChatID:    msg.ChatId,
		UserID:    msg.UserId,
		Text:      replyText,
		Direction: "outgoing",
		TimeStamp: time.Now(),
	}

	// Сохранаяем сообщение в БД через сервисный слой
	err = b.Service.CheckAndSaveMsg(ctx, outgoingMsg)
	if err != nil {
		log.Printf("failed to save outgoing message: %v", err)
		//return nil, fmt.Errorf("failed to save outgoing message: %w", err)
	}

	// в ответе пеперадём юзеру тестовую кливиатуру (через grpc на сервер бота)
	replyMakrUp := converter.ToProtoReplyMarkup(b.Service.CreateWelcomeReplyKeyboard())

	// Формируем ответ для бота

	response := &pb.UpdateResponse{
		Success: true,
		Messages: []*pb.OutgoingMessage{
			{
				ChatId:      msg.ChatId,
				Text:        replyText,
				ReplyMarkup: replyMakrUp,
			},
		},
	}

	return response, nil
}

// ProcessCallback - обработка callback от inline клавиатуры
func (b *BizGRPCHandler) ProcessCallback(ctx context.Context, callback *pb.CallbackQuery) (*pb.UpdateResponse, error) {
	// Проверяем, что callback не nil и содержит необходимые данные
	if callback == nil {
		return nil, fmt.Errorf("callback is nil")
	}

	// Добавим подробное логирование для отладки
	log.Printf("🔍 ProcessCallback получил: ID=%s, UserID=%d, ChatID=%d, MessageID=%d, Data=%s",
		callback.Id, callback.UserId, callback.ChatId, callback.MessageId, callback.Data)

	// Приводим входящий callback к domain типу, используем конвертер
	callBackLog := converter.ToCallbackLog(callback)

	log.Println(callBackLog.MessageID)

	// при вызове этого метода необходимо сохранить нового пользователя в базе или обновить некоторые его поля
	user, err := b.Service.RegisterOrUpdate(ctx, callBackLog.MessageID, callBackLog.UserFirstName, callBackLog.UserLastName, callBackLog.UserNickName)
	if err != nil {
		log.Printf("failed to save/update user: %v", err)
	}

	fmt.Println("-----------------------------------------")
	fmt.Println(user)
	fmt.Println("-----------------------------------------")

	// логируем, что пользователь нажал на кнопку
	log.Printf("Пользователь %s нажал на кнопку, началась обработка колбэка %v \n", user.Username, callBackLog.ID)

	// проверяем и сохраняем данные в БД
	err = b.Service.CheckAndSaveCallBack(ctx, callBackLog)
	if err != nil {
		log.Printf("⚠️ failed to save callback: %v", err) // логируем, но не прерываем
	}

	// Анализируем данные callback и формируем ответ
	response := &pb.UpdateResponse{
		Success: true,
	}

	// ИСПОЛЬЗУЕМ callback.ChatId - это правильное поле!
	switch callback.Data {
	case "help":
		response.Messages = append(response.Messages, &pb.OutgoingMessage{
			ChatId: callback.ChatId,
			Text:   "Я бот-помощник. Доступные команды:\n/help - помощь\n",
		})

	case "lookup":
		response.Messages = append(response.Messages, &pb.OutgoingMessage{
			ChatId: callback.ChatId,
			Text:   "Вот ссылка на инстграмм аккаунт мастера:\nhttps://www.google.com\n",
		})

	default:
		response.Messages = append(response.Messages, &pb.OutgoingMessage{
			ChatId: callback.ChatId,
			Text:   fmt.Sprintf("Неизвестная команда: %s", callback.Data),
		})
	}

	log.Printf("✅ ProcessCallback отправляет ответ с %d сообщениями", len(response.Messages))
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
