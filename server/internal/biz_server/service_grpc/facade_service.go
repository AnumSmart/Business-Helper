package servicegrpc

import (
	"server/internal/biz_server/grpcclient"
	"server/internal/biz_server/repository"
)

// общая структура для GRPC сервиса
type BizServiceFacade struct {
	Users     UserService
	Messages  MessageService
	Responses ResponseGenerator
}

// конструктор для GRPC сервиса
func NewBizServiceFacade(repo *repository.BizRepository, grpcClient *grpcclient.BotGrpcClient) *BizServiceFacade {
	return &BizServiceFacade{
		Users:     NewUserService(repo),
		Messages:  NewMessageService(repo, grpcClient),
		Responses: NewResponseGenerator(),
	}
}
