package httpserver

import (
	"bot/internal/config"
	"bot/internal/server/http_server/handlers"
	"context"
	"fmt"
	"log"
	"net/http"
	"pkg/middleware"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	tele "gopkg.in/telebot.v4"
)

// структура серверя для ботов
type BotGateway struct {
	httpServer *http.Server                // базовый сервер из пакета http
	router     *gin.Engine                 // роутер gin
	config     *config.BotHttpServerConfig // конфиг http сервера на базе общего конфига
	botConfig  *config.BotConfig           // конфиг бота
	Handler    *handlers.BotHttpHandler    // хэндлер
	stopChan   chan struct{}               // канал для синхронизации горутин

	// Добавляем поля для Telegram бота (этот бот будет использоваться только в longpolling режиме)
	telegramBot *tele.Bot          // экземпляр бота
	botWg       sync.WaitGroup     // для ожидания завершения бота (бот будет запускаться в отдельной горутине, это блокирующая операция)
	botCtx      context.Context    // контекст для управления ботом
	botCancel   context.CancelFunc // функция отмены для бота
}

// Конструктор для сервера
func NewBotGateway(ctx context.Context, config *config.BotHttpServerConfig, botConf *config.BotConfig, handler *handlers.BotHttpHandler) (*BotGateway, error) {
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
		router:    router,
		config:    config,
		botConfig: botConf,
		Handler:   handler,
		stopChan:  make(chan struct{}),
	}, nil
}

// Метод для маршрутизации сервера при режиме webhook
func (a *BotGateway) SetUpWebHookRoutes() {
	a.router.POST("/webhook", a.Handler.HandleWebhook) // основной метод, если в конфиге прописан режим webhook
}

// Метод для настройки и запуска long polling бота
func (a *BotGateway) SetUpPollingRoutes() error {
	// Создаём контекст для управления ботом
	a.botCtx, a.botCancel = context.WithCancel(context.Background())

	// Настройки бота из конфига (предполагаю, что у вас есть поля в конфиге)
	pref := tele.Settings{
		Token:  a.botConfig.BotToken,                        // добавьте это поле в ваш конфиг
		Poller: &tele.LongPoller{Timeout: 30 * time.Second}, // интервал запросов к телеграмм на обновления
	}

	// Создаём бота
	bot, err := tele.NewBot(pref)
	if err != nil {
		return fmt.Errorf("Error during construction of polling bot:%v\n", err)
	}

	// назначаем этого polling бота в структуру сервера
	a.telegramBot = bot

	// Регистрируем обработчики бота, передавая управление Handler слой сервера
	a.registerBotHandlers()

	// Запускаем бота асинхронно
	a.botWg.Add(1) // добавляем 1 горутину в вэйт группу
	go a.runBot()  // запускаем бот в отдельной горутине

	log.Println("Long polling бот успешно запущен в фоновом режиме")
	return nil
}

// метод для связывания сообщения от Telegram с бизнес-логикой через Handler
func (a *BotGateway) registerBotHandlers() {
	// Обработка callback-запросов от inline клавиатур
	a.telegramBot.Handle(tele.OnCallback, func(c tele.Context) error {
		return a.Handler.HandleBotCallback(c)
	})

	// Обработка команды /start
	a.telegramBot.Handle("/start", func(c tele.Context) error {
		// Передаём управление в ваш handler
		return a.Handler.HandleBotStart(c)
	})

	// Обработка всех текстовых сообщений
	a.telegramBot.Handle(tele.OnText, func(c tele.Context) error {
		return a.Handler.HandleBotMessage(c)
	})

}

// метод для запуска polling бота
func (a *BotGateway) runBot() {
	defer a.botWg.Done()

	log.Println("Telegram bot (long polling) started")

	// Запускаем бота. Start() блокируется, поэтому мы в горутине
	go func() {
		a.telegramBot.Start()
	}()

	// Ожидаем сигнала завершения
	select {
	case <-a.botCtx.Done():
		log.Println("Получен сигнал остановки бота")
	case <-a.stopChan:
		log.Println("Получен сигнал остановки сервера")
	}

	// Останавливаем бота корректно
	a.telegramBot.Stop()
	log.Println("Telegram bot (long polling) stopped")
}

// Метод для запуска сервера
func (a *BotGateway) Run() error {
	switch a.config.Mode {
	case "webhook":
		a.SetUpWebHookRoutes()
	case "polling":
		if err := a.SetUpPollingRoutes(); err != nil {
			return err
		}
	}

	a.httpServer = &http.Server{
		Handler: a.router,
	}
	// Используем обычный порт для HTTP
	a.httpServer.Addr = a.config.Addr()
	log.Printf("Starting HTTP server on %s in %s mode", a.config.Addr(), a.config.Mode)
	return a.httpServer.ListenAndServe()
}

// Метод для graceful shutdown
func (a *BotGateway) Shutdown(ctx context.Context) error {
	log.Println("Начинаем graceful shutdown...")

	// 1️⃣ Сначала закрываем HTTP сервер (перестаем принимать новые запросы)
	// Это важно сделать первым, чтобы новые запросы не пошли в уже закрывающиеся клиенты
	if err := a.httpServer.Shutdown(ctx); err != nil {
		return err
	}

	// 2️⃣ Если бот запущен в polling режиме, останавливаем его
	if a.telegramBot != nil {
		log.Println("Останавливаем Telegram бота...")
		a.botCancel() // Отправляем сигнал остановки

		// Ждём завершения с таймаутом
		done := make(chan struct{})
		go func() {
			a.botWg.Wait()
			close(done)
		}()

		select {
		case <-done:
			log.Println("Telegram бот остановлен")
		case <-time.After(5 * time.Second):
			log.Println("Таймаут при остановке бота")
		}
	}

	// 3️⃣ Сигнализируем всем горутинам о завершении
	close(a.stopChan)

	//  Даем время завершить текущие операции (например, отправку сообщений)
	time.Sleep(1 * time.Second)

	//  Для Telegram клиента - просто логируем (можно ничего не делать)
	log.Println("Telegram клиент: ресурсы будут очищены сборщиком мусора")

	log.Println("HTTP BOT Server shutdown completed")
	return nil
}
