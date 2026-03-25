package handlersgrpc

import (
	"server/internal/biz_server/service"
	"server/internal/interfaces"
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
