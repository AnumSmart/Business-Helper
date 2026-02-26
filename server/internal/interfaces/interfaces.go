package interfaces

import (
	"context"
	pb "global_models/grpc/bot"
)

// интерфейс слоя базы данных для авторизации юзеров
type DBRepoInterface interface {
}

// интерфейс кэша
type CacheRepoInterface interface {
}

// интерфейс работы хэндлера по grpc (констракты согласно .proto)
type GRPCHandlerInterface interface {
	// ProcessMessage - обработка входящего сообщения
	ProcessMessage(ctx context.Context, msg *pb.Message) (*pb.UpdateResponse, error)

	// ProcessCallback - обработка callback от inline клавиатуры
	ProcessCallback(ctx context.Context, callback *pb.CallbackQuery) (*pb.UpdateResponse, error)

	//ProcessIncomingMsg - обработка сообщения
	ProcessIncomingMsg(ctx context.Context, req *pb.SendMessageRequest) (*pb.SendMessageResponse, error)
}
