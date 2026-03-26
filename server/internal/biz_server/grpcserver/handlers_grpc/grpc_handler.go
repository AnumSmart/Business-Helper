package handlersgrpc

import (
	servicegrpc "server/internal/biz_server/service_grpc"
	"server/internal/interfaces"
)

// На этом слое остается только транспортная логика (преобразование данных и управление запросом/ответом)
type BizGRPCHandler struct {
	Service *servicegrpc.BizServiceFacade
}

func NewBizGRPCHandler(grpcService *servicegrpc.BizServiceFacade) interfaces.GRPCHandlerInterface {
	return &BizGRPCHandler{
		Service: grpcService,
	}
}
