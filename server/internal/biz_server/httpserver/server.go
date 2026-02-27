package httpserver

import (
	"context"
	"global_models/interf"
	"log"
	"net/http"
	"pkg/configs"
	"pkg/middleware"

	"github.com/gin-gonic/gin"
)

// структура сервера для управления логикой ботов-ассистентов
type BizServer struct {
	httpServer *http.Server                   // базовый сервер из пакета http
	router     *gin.Engine                    // роутер gin
	config     *configs.ServerConfig          // базовый конфиг
	Handler    interf.BizHTTPHandlerInterface // интерфейс слоя хэндлеров
}

// Конструктор для сервера
func NewBizServer(ctx context.Context, config *configs.ServerConfig, handler interf.BizHTTPHandlerInterface) (*BizServer, error) {
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

	return &BizServer{
		router:  router,
		config:  config,
		Handler: handler,
	}, nil
}

// Метод для маршрутизации сервера
func (a *BizServer) SetUpRoutes() {
	a.router.GET("/echo", a.Handler.EchoServer) // тестовый ендпоинт
}

// Метод для запуска сервера
func (a *BizServer) Run() error {
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
func (a *BizServer) Shutdown(ctx context.Context) error {

	// Останавливаем HTTP сервер
	if err := a.httpServer.Shutdown(ctx); err != nil {
		return err
	}

	log.Println("Server shutdown completed")
	return nil
}
