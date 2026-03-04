package dependencies

import (
	"bot/configs"
	"bot/internal/config"
	grpcclient "bot/internal/server/grpc_client"
	handlersgrpc "bot/internal/server/grpc_server/handlers_grpc"
	"bot/internal/server/http_server/handlers"
	"bot/internal/server/service"
	"sync"

	httpclient "bot/internal/server/http_client"
	"context"

	"fmt"
	"runtime"
)

// определяем зависимости для сервиса ботов
type BotServiceDependencies struct {
	BotConfig       *config.BotConfig            // структрура конфига для создания бота
	BotServerconfig *configs.BotServiceConfig    // конфиг сервиса
	BotGrpcClient   *grpcclient.BotGrpcClient    // клиент для работы по grpc
	BotHTTPClient   *httpclient.BotHTTPClient    // клиент для работы по HTTP
	BotHttpHandler  *handlers.BotHttpHandler     // хэндлер для http сервера бота
	BotGrpcHandler  *handlersgrpc.BotGRPCHandler // хэндлер для grpc сервера бота

	closeOnce sync.Once // для того, чтобы функция освобождения ресурсов выполнилась только 1 раз
	closeErr  error
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
	botGrpcClient, err := grpcclient.NewBotGrpcClient(botConf.GRPCServer)
	if err != nil {
		return nil, fmt.Errorf("failed to create botGRPC client: %w", err)
	}

	// создаём клиент, который может общаться по HTTP
	botHTTPClient := httpclient.NewClient(botConf.BotToken)

	// создаём сервисный слой для бота
	botService := service.NewBotService(botGrpcClient, botHTTPClient)

	// создаём хэндлер для http сервера бота
	botHttpHandler := handlers.NewBotHandler(botService)

	// сощдаём хэндлер для grpc сервера бота
	botGrpcHandler := handlersgrpc.NewBotGRPCHandler(botService)

	return &BotServiceDependencies{
		BotConfig:       botConf,
		BotServerconfig: serviceConf,
		BotGrpcClient:   botGrpcClient,
		BotHTTPClient:   botHTTPClient,
		BotHttpHandler:  botHttpHandler,
		BotGrpcHandler:  botGrpcHandler,
	}, nil
}

// метод структуры зависимостей для осбобождения ресурсов
func (d *BotServiceDependencies) Close() error {
	d.closeOnce.Do(func() {
		var errs []error

		// Закрываем gRPC клиент (В ПЕРВУЮ ОЧЕРЕДЬ!)
		if d.BotGrpcClient != nil {
			if err := d.BotGrpcClient.Close(); err != nil {
				errs = append(errs, fmt.Errorf("grpc client: %w", err))
			}
		}

		// проверяем аггрегированные ошибки
		if len(errs) > 0 {
			d.closeErr = fmt.Errorf("close errors: %v", errs)
		}
	})

	// выводим сообщение, что ресурсы освобождены
	if d.closeErr == nil {
		fmt.Println("Ресурсы - освобождены")
	}

	return d.closeErr
}
