package interf

import (
	"context"
	pb "global_models/grpc/bot"

	"github.com/gin-gonic/gin"
)

// общий интерфейс работы по GRPC
type GRPCInterface interface {
	ProcessUpdate(ctx context.Context, req *pb.UpdateRequest) (*pb.UpdateResponse, error)
	SendMessage(ctx context.Context, req *pb.SendMessageRequest) (*pb.SendMessageResponse, error)
	Shutdown(ctx context.Context) error
}

type BizHTTPHandlerInterface interface {
	EchoServer(c *gin.Context)
}
