package servicegrpc

import (
	"context"
	"fmt"
	"server/internal/biz_server/grpcclient"
	"server/internal/biz_server/repository"
	"server/internal/domain"
)

// ========== Message Service ==========
type MessageService interface {
	CheckAndSaveMsg(ctx context.Context, msg *domain.Message) error
	CheckAndSaveCallBack(ctx context.Context, callBackLog *domain.CallbackLog) error
	ProcessIncomingMessage(ctx context.Context, req *domain.IncomingMessage) (*domain.MessageResponse, error)
}

// структура сервиса сообщений
type messageService struct {
	repo       *repository.BizRepository
	grpcClient *grpcclient.BotGrpcClient
}

// конструктор сервиса вообщений
func NewMessageService(repo *repository.BizRepository, grpcClient *grpcclient.BotGrpcClient) MessageService {
	return &messageService{
		repo:       repo,
		grpcClient: grpcClient,
	}
}

// метод для проверки и сохданения входящего сообщения (по GRPC) в базу
func (s *messageService) CheckAndSaveMsg(ctx context.Context, msg *domain.Message) error {
	if msg == nil {
		return fmt.Errorf("Incoming messgage can not be nil! [error in service layer]")
	}

	return s.repo.Save(ctx, msg)
}

func (s *messageService) CheckAndSaveCallBack(ctx context.Context, callBackLog *domain.CallbackLog) error {
	if callBackLog == nil {
		return fmt.Errorf("callback log can not be nil")
	}

	return s.repo.SaveCallback(ctx, callBackLog)
}

func (s *messageService) ProcessIncomingMessage(ctx context.Context, req *domain.IncomingMessage) (*domain.MessageResponse, error) {
	// Здесь может быть валидация, сохранение в БД, etc.
	return &domain.MessageResponse{
		Success: true,
	}, nil
}
