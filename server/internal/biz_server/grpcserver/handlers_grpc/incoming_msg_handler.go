package handlersgrpc

import (
	"context"
	"fmt"
	pb "global_models/grpc/bot"
	"server/internal/biz_server/grpcserver/converter"
)

// Принятое сообщение от grpc клиента обрабатывается в слое хэндлера и передаётся в слой сервиса
func (b *BizGRPCHandler) ProcessIncomingMsg(ctx context.Context, req *pb.SendMessageRequest) (*pb.SendMessageResponse, error) {
	// 1. Валидация
	if err := b.validateRequest(req); err != nil {
		return nil, err
	}

	// 2. Конвертация и вызов сервиса
	incomingMsg := converter.ToIncomingMessage(req)
	result, err := b.Service.AnswerIncomingMsg(ctx, incomingMsg)
	if err != nil {
		return nil, fmt.Errorf("service error: %w", err)
	}

	// 3. Конвертация ответа
	return converter.ToProtoResponse(result), nil
}

func (b *BizGRPCHandler) validateRequest(req *pb.SendMessageRequest) error {
	if req == nil {
		return fmt.Errorf("request is nil")
	}
	if req.ChatId == 0 {
		return fmt.Errorf("chat ID must not be 0")
	}
	return nil
}
