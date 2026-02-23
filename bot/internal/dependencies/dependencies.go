package dependencies

import (
	"bot/configs"
	"bot/internal/config"
	grpc_client "bot/internal/server/grpc_client"
	"bot/internal/server/handlers"
	httpclient "bot/internal/server/http_client"
	"context"

	"fmt"
	"runtime"
)

// определяем зависимости для сервиса ботов
type BotServiceDependencies struct {
	BotConfig       *config.BotConfig          // структрура конфига для создания бота
	BotServerconfig *configs.BotServiceConfig  // базовый конфиг
	BotGrpcClient   *grpc_client.BotGrpcClient // клиент для работы по grpc
	BotHTTPClient   *httpclient.BotHTTPClient  // клиент для работы по HTTP
	BotHandler      *handlers.BotHandler       // хэндлер для сервиса ботов
}

// InitDependencies инициализирует общие зависимости для bot_service
func InitDependencies(ctx context.Context) (*BotServiceDependencies, error) {
	// Получаем количество CPU
	currentMaxProcs := runtime.GOMAXPROCS(-1)
	fmt.Printf("Текущее значение GOMAXPROCS: %d\n", currentMaxProcs)

	// Получаем конфигурацию (сервиса бота)
	serviceConf, err := configs.LoadBotServiceConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to load service config: %w", err)
	}

	// получаем конфигурацию бота
	botConf, err := config.LoadBotConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to load bot config: %w", err)
	}

	// создаём клиент, который может общаться по grpc
	botGrpcClient, err := grpc_client.NewBotGrpcClient("localhost:50051") // пока жестко забили адрес, нужно вынести в конфиг!
	if err != nil {
		return nil, fmt.Errorf("failed to create botGRPC client: %w", err)
	}

	fmt.Println("создали grpc клиент")

	// создаём клиент, который может общаться по HTTP
	botHTTPClient := httpclient.NewClient(botConf.BotToken)

	// создаём хэндлер для сервиса бота
	botHandler := handlers.NewBotHandler(botGrpcClient, botHTTPClient)

	return &BotServiceDependencies{
		BotConfig:       botConf,
		BotServerconfig: serviceConf,
		BotGrpcClient:   botGrpcClient,
		BotHTTPClient:   botHTTPClient,
		BotHandler:      botHandler,
	}, nil
}
