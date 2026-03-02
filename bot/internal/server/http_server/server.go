package httpserver

import (
	"bot/internal/server/http_server/handlers"
	"context"
	"log"
	"net/http"
	"pkg/configs"
	"pkg/middleware"
	"time"

	"github.com/gin-gonic/gin"
)

// структура серверя для ботов
type BotGateway struct {
	httpServer *http.Server              // базовый сервер из пакета http
	router     *gin.Engine               // роутер gin
	config     *configs.HttpServerConfig // базовый конфиг
	Handler    *handlers.BotHttpHandler  // хэндлер
	stopChan   chan struct{}             // канал для синхронизации горутин
}

// Конструктор для сервера
func NewBotGateway(ctx context.Context, config *configs.HttpServerConfig, handler *handlers.BotHttpHandler) (*BotGateway, error) {
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

	return &BotGateway{
		router:   router,
		config:   config,
		Handler:  handler,
		stopChan: make(chan struct{}),
	}, nil
}

// Метод для маршрутизации сервера
func (a *BotGateway) SetUpRoutes() {
	a.router.POST("/webhook", a.Handler.HandleWebhook) // основной метод
}

// Метод для запуска сервера
func (a *BotGateway) Run() error {
	a.SetUpRoutes()

	a.httpServer = &http.Server{
		Handler: a.router,
	}
	// Используем обычный порт для HTTP
	a.httpServer.Addr = a.config.Addr()
	log.Printf("Starting HTTP server on %s", a.config.Addr())
	return a.httpServer.ListenAndServe()
}

// Метод для graceful shutdown
func (a *BotGateway) Shutdown(ctx context.Context) error {

	// 1️⃣ Сначала закрываем HTTP сервер (перестаем принимать новые запросы)
	// Это важно сделать первым, чтобы новые запросы не пошли в уже закрывающиеся клиенты
	if err := a.httpServer.Shutdown(ctx); err != nil {
		return err
	}

	// 2️⃣ Сигнализируем всем горутинам о завершении
	close(a.stopChan)

	// 3️⃣ Даем время завершить текущие операции (например, отправку сообщений)
	time.Sleep(1 * time.Second)

	// 5️⃣ Для Telegram клиента - просто логируем (можно ничего не делать)
	log.Println("Telegram клиент: ресурсы будут очищены сборщиком мусора")

	log.Println("HTTP BOT Server shutdown completed")
	return nil
}
