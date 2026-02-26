package dependencies

import (
	"context"
	"fmt"
	"global_models/global_cache"
	"global_models/global_db"
	postgresdb "pkg/postgres_db"
	"pkg/redis"
	"runtime"

	"server/configs"
	"server/internal/biz_server/grpcclient"
	handlersgrpc "server/internal/biz_server/grpcserver/handlers_grpc"
	"server/internal/biz_server/httpserver/handlers"
	"server/internal/biz_server/repository"
	"server/internal/biz_server/service"
	"server/internal/interfaces"
	"sync"
)

// Dependencies содержит все общие зависимости
type BizServiceDepenencies struct {
	BizConfig      *configs.BizServiceConfig        // конфиг всего сервера управления ботами
	BizHTTPHandler handlers.BizHTTPHandlerInterface // интерфейс хэндлера
	BizGRPCHandler interfaces.GRPCHandlerInterface  // интерфейс хэндлера для работы по grpc

	// добавляем поля для логики освобождения ресурсов
	pgPool         global_db.Pool     // для особождения ресурсов DB
	redisCacherepo global_cache.Cache // для освобождения ресурсов redis
	closeOnce      sync.Once          // для того, чтобы функция освобождения ресурсов выполнилась только 1 раз
	closeErr       error
}

// InitDependencies инициализирует общие зависимости для auth_service
func InitDependencies(ctx context.Context) (*BizServiceDepenencies, error) {
	// Получаем количество CPU
	currentMaxProcs := runtime.GOMAXPROCS(-1)
	fmt.Printf("Текущее значение GOMAXPROCS: %d\n", currentMaxProcs)

	// Получаем конфигурацию
	conf, err := configs.LoadConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	// создаём экземпляр пула соединений для postgresQL
	// адаптер к глобальному интерфейсу используется внутри NewPoolWithConfig
	pgPool, err := postgresdb.NewPoolWithConfig(ctx, conf.PostgresDBConf)
	if err != nil {
		return nil, fmt.Errorf("failed to create PostgreSQL repository: %w", err)
	}

	// создаём репозиторий для авторизации пользователя
	bizRepo := repository.NewBizDBRepository(pgPool)

	// создаём экземпляр redis
	redisCacherepo, err := redis.NewRedisCacheRepository(conf.RedisConf)
	if err != nil {
		return nil, fmt.Errorf("failed to create Black List repository (based om Redis): %w", err)
	}

	// создаём репозиторий кэша
	cache, err := repository.NewBizCacheRepo(redisCacherepo, "server_cache")

	// создаём слой репозитория (на базе репозитория Postgres и кэша (на базе redis))
	repo, err := repository.NewBizRepository(bizRepo, cache)
	if err != nil {
		return nil, fmt.Errorf("failed to create Auth Repository Layer: %w", err)
	}

	// создаём экземпляр grpc клиента
	grpcClient, err := grpcclient.NewBotGrpcClient(conf.GRPCClientConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create grpc client: %w", err)
	}

	// создаём сервисный слой
	service, err := service.NewBizService(repo, grpcClient)
	if err != nil {
		return nil, fmt.Errorf("failed to create Service Layer: %w", err)
	}

	// создаём слой хэндлера для HTTP
	bizHTTPHandler := handlers.NewBizHandler(service)
	if bizHTTPHandler == nil {
		return nil, fmt.Errorf("failed to create bizness http handler")
	}

	// создаём слой хэндлера для GRPC
	bizGRPCHandler := handlersgrpc.NewBizGRPCHandler(service)
	if bizGRPCHandler == nil {
		return nil, fmt.Errorf("failed to create bizness grpc handler")
	}

	return &BizServiceDepenencies{
		BizConfig:      conf,
		BizHTTPHandler: bizHTTPHandler,
		BizGRPCHandler: bizGRPCHandler,
	}, nil
}

// метод структуры зависимостей для осбобождения ресурсов
func (d *BizServiceDepenencies) Close() error {
	d.closeOnce.Do(func() {
		var errs []error

		// Закрываем Redis
		if d.redisCacherepo != nil {
			if err := d.redisCacherepo.Close(); err != nil {
				errs = append(errs, fmt.Errorf("redis: %w", err))
			}
		}

		// Закрываем PostgreSQL
		if d.pgPool != nil {
			if err := d.pgPool.Close(); err != nil {
				errs = append(errs, fmt.Errorf("postgres: %w", err))
			}
		}

		// проверяем аггрегированные ошибки
		if len(errs) > 0 {
			d.closeErr = fmt.Errorf("close errors: %v", errs)
		}
	})

	if d.closeErr == nil {
		fmt.Println("Ресурсы - освобождены")
	}

	// если все хоршо, то возвращаем nil
	return d.closeErr
}
