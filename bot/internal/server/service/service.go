package service

import (
	grpcclient "bot/internal/server/grpc_client"
	httpclient "bot/internal/server/http_client"
	"context"
	"fmt"

	pb "global_models/grpc/bot"
)

// структура сервисного слоя бота
type BotService struct {
	grpcClient *grpcclient.BotGrpcClient // Для отправки данных в gRPC сервер
	hTTPClient *httpclient.BotHTTPClient // Для отправки ответов в Telegram
}

// конструктор для создания сервисного слоя бота
func NewBotService(grpcClient *grpcclient.BotGrpcClient, tgClient *httpclient.BotHTTPClient) *BotService {
	return &BotService{
		grpcClient: grpcClient,
		hTTPClient: tgClient,
	}
}

// метод сервисного слоя бота для обработки обновелния от телеграмм и отправки ответа
func (b *BotService) ProcessUpdate(ctx context.Context, req *pb.UpdateRequest) (*pb.UpdateResponse, error) {
	resp, err := b.grpcClient.ProcessUpdate(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("Ошибка связи с gRPC сервером")
	}

	return resp, nil
}

// метод сервисного слоя бота для отправки обработанных сообщений по http
func (b *BotService) SendHTTPMessages(msgs []*pb.OutgoingMessage) error {
	return b.hTTPClient.SendOutgoingMessages(msgs)
}
