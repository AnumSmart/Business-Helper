package server

import (
	"bot/internal/server/handlers"
	"context"
	"fmt"
	"log"
	"net/http"
	"pkg/configs"
	"pkg/middleware"
	"time"

	"github.com/gin-gonic/gin"
)

// структура серверя для ботов
type BotServiceServer struct {
	httpServer *http.Server          // базовый сервер из пакета http
	router     *gin.Engine           // роутер gin
	config     *configs.ServerConfig // базовый конфиг
	Handler    *handlers.BotHandler  // хэндлер
	stopChan   chan struct{}         // канал для синхронизации горутин
}

// Конструктор для сервера
func NewBotServiceServer(ctx context.Context, config *configs.ServerConfig, handler *handlers.BotHandler) (*BotServiceServer, error) {
	// создаём экземпляр роутера
	router := gin.Default()
	err := router.SetTrustedProxies(nil)
	if err != nil {
		return nil, err
	}

	// Добавляем middleware для проброса контекста
	router.Use(func(c *gin.Context) {
		ctx := context.WithValue(c.Request.Context(), "request_id", c.GetHeader("X-Request-ID"))
		c.Request = c.Request.WithContext(ctx)
		c.Next()
	})

	router.Use(middleware.CORSMiddleware()) // используем для всех маршруторв работу с CORS

	return &BotServiceServer{
		router:   router,
		config:   config,
		Handler:  handler,
		stopChan: make(chan struct{}),
	}, nil
}

// Метод для маршрутизации сервера
func (a *BotServiceServer) SetUpRoutes() {
	a.router.POST("/webhook", a.Handler.HandleWebhook) // основной метод
}

// Метод для запуска сервера
func (a *BotServiceServer) Run() error {
	a.SetUpRoutes()
	fmt.Println("установили роуты!")

	a.httpServer = &http.Server{
		Handler: a.router,
	}
	// Используем обычный порт для HTTP
	a.httpServer.Addr = a.config.Addr()
	log.Printf("Starting HTTP server on %s", a.config.Addr())
	return a.httpServer.ListenAndServe()
}

// Метод для graceful shutdown
func (a *BotServiceServer) Shutdown(ctx context.Context) error {

	// 1️⃣ Сначала закрываем HTTP сервер (перестаем принимать новые запросы)
	// Это важно сделать первым, чтобы новые запросы не пошли в уже закрывающиеся клиенты
	if err := a.httpServer.Shutdown(ctx); err != nil {
		return err
	}

	// 2️⃣ Сигнализируем всем горутинам о завершении
	close(a.stopChan)

	// 3️⃣ Даем время завершить текущие операции (например, отправку сообщений)
	time.Sleep(1 * time.Second)

	// 4️⃣ Закрываем gRPC клиент (активное соединение)
	if a.Handler.GrpcClient != nil {
		log.Println("📞 Закрываем gRPC соединение...")
		if err := a.Handler.GrpcClient.Close(); err != nil {
			log.Printf("Ошибка при закрытии gRPC клиента: %v", err)
		}
	}

	// 5️⃣ Для Telegram клиента - просто логируем (можно ничего не делать)
	log.Println("Telegram клиент: ресурсы будут очищены сборщиком мусора")

	log.Println("Server shutdown completed")
	return nil
}
