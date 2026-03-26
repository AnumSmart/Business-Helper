package handlersgrpc

import (
	"context"
	"fmt"
	pb "global_models/grpc/bot"
	"server/internal/biz_server/grpcserver/converter"
	"server/internal/domain"
	"time"
)

// объединяет данные для обработки сообщения
type messageContext struct {
	ctx    context.Context
	msg    *domain.Message
	user   *domain.User
	chatID int64
	userID int64
}

// ProcessMessage - обработка входящего сообщения
func (b *BizGRPCHandler) ProcessMessage(ctx context.Context, msg *pb.Message) (*pb.UpdateResponse, error) {
	// 1. Создание контекста сообщения
	msgCtx, err := b.buildMessageContext(ctx, msg)
	if err != nil {
		return nil, err
	}

	// 2. Логирование входящего сообщения
	b.logIncomingMessage(msg)

	// 3. Сохранение пользователя и сообщения
	if err := b.saveUserAndMessage(msgCtx); err != nil {
		fmt.Printf("⚠️ Warning: %v", err)
		// продолжаем выполнение, не блокируем ответ
	}

	// 4. Генерация ответа
	replyText := b.Service.Responses.GenerateReply(msg.Text, msgCtx.user)
	replyMarkup := b.Service.Responses.CreateTextRespKeyBoard(msg.Text)

	// 5. Сохранение исходящего сообщения
	b.saveOutgoingMessage(msgCtx.ctx, msgCtx.chatID, msgCtx.userID, replyText)

	// 6. Формирование ответа
	return b.buildMessageResponse(msgCtx.chatID, replyText, replyMarkup), nil
}

// строим контекст сообщения на базе внутренней структуры
func (b *BizGRPCHandler) buildMessageContext(ctx context.Context, msg *pb.Message) (*messageContext, error) {
	incomingMsg := converter.ToDomainMessage(msg)
	user := converter.ToDomainUser(msg.From)

	return &messageContext{
		ctx:    ctx,
		msg:    incomingMsg,
		user:   user,
		chatID: msg.ChatId,
		userID: msg.UserId,
	}, nil
}

// логируем сообщение (пока в консоль)
func (b *BizGRPCHandler) logIncomingMessage(msg *pb.Message) {
	fmt.Printf("📨 Incoming message: UserID=%d, ChatID=%d, Text=%s",
		msg.UserId, msg.ChatId, msg.Text)
}

// метод для сохранения/обновления пользователя и его сообщения
func (b *BizGRPCHandler) saveUserAndMessage(msgCtx *messageContext) error {
	// Сохраняем/обновляем пользователя
	_, err := b.Service.Users.RegisterOrUpdate(msgCtx.ctx,
		msgCtx.msg.MessageID,
		msgCtx.msg.UserFirstName,
		msgCtx.msg.UserLastName,
		msgCtx.msg.UserNickName)
	if err != nil {
		return fmt.Errorf("failed to save user: %w", err)
	}

	// Сохраняем сообщение
	if err := b.Service.Messages.CheckAndSaveMsg(msgCtx.ctx, msgCtx.msg); err != nil {
		return fmt.Errorf("failed to save message: %w", err)
	}

	return nil
}

// метод для сохранения исходящего сообщения (на базе доменной структуры)
func (b *BizGRPCHandler) saveOutgoingMessage(ctx context.Context, chatID, userID int64, text string) {
	outgoingMsg := &domain.Message{
		ChatID:    chatID,
		UserID:    userID,
		Text:      text,
		Direction: "outgoing",
		TimeStamp: time.Now(),
	}

	if err := b.Service.Messages.CheckAndSaveMsg(ctx, outgoingMsg); err != nil {
		fmt.Printf("⚠️ Failed to save outgoing message: %v", err)
	}
}

// метод для формирования ответа в grpc форме
func (b *BizGRPCHandler) buildMessageResponse(chatID int64, text string, markup *domain.ReplyMarkup) *pb.UpdateResponse {
	return &pb.UpdateResponse{
		Success: true,
		Messages: []*pb.OutgoingMessage{
			{
				ChatId:      chatID,
				Text:        text,
				ReplyMarkup: converter.ToProtoReplyMarkup(markup),
			},
		},
	}
}
