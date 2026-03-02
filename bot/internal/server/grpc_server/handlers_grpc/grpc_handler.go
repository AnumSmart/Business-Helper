package handlersgrpc

import "bot/internal/server/service"

// На этом слое остается только транспортная логика (преобразование данных и управление запросом/ответом)
type BotGRPCHandler struct {
	Service *service.BotService
}

// конструктор для слоя хэндлеров grpc сервера бота
func NewBotGRPCHandler(service *service.BotService) *BotGRPCHandler {
	return &BotGRPCHandler{
		Service: service,
	}
}
