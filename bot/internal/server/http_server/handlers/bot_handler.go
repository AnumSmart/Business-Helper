package handlers

import (
	"bot/internal/domain"
	"bot/internal/server/http_server/converter"
	"bot/internal/server/service"
	"net/http"

	"github.com/gin-gonic/gin"
)

// BotHttpHandler обрабатывает входящие вебхуки от Telegram
type BotHttpHandler struct {
	BotService *service.BotService
}

// NewBotHandler создает новый экземпляр обработчика с внедренными зависимостями
// Паттерн "Dependency Injection" - клиенты передаются извне
func NewBotHandler(botService *service.BotService) *BotHttpHandler {
	return &BotHttpHandler{
		BotService: botService,
	}
}

// HandleWebhook - основной метод обработки входящих вебхуков от Telegram
// Принимает gin.Context для доступа к запросу и ответу
func (h *BotHttpHandler) HandleWebhook(c *gin.Context) {
	var update domain.TelegramUpdate

	// ShouldBindJSON автоматически парсит JSON из тела запроса в структуру
	// Возвращает ошибку, если JSON невалиден или не соответствует структуре
	if err := c.ShouldBindJSON(&update); err != nil {
		// В случае ошибки возвращаем 400 Bad Request
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Шаг 1: Конвертируем Telegram формат в gRPC формат
	grpcUpdate := converter.ConvertToGRPCUpdate(&update)

	// Шаг 2: Отправляем на gRPC сервер для бизнес-логики
	// c.Request.Context() передает контекст HTTP запроса в gRPC вызов
	resp, err := h.BotService.ProcessUpdate(c.Request.Context(), grpcUpdate)
	if err != nil {
		// Ошибка связи с gRPC сервером
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Шаг 3: Если сервер вернул сообщения для отправки - отправляем их в Telegram
	if resp.Success && len(resp.Messages) > 0 {
		if err := h.BotService.SendHTTPMessages(resp.Messages); err != nil {
			// Важно: даже если не удалось отправить ответ, мы не возвращаем ошибку Telegram
			// Иначе Telegram будет повторно отправлять тот же update
			c.JSON(http.StatusOK, gin.H{"status": "processed but failed to send response"})
			return
		}
	}

	// Успешная обработка - возвращаем 200 OK
	// Telegram ожидает 200, чтобы не переотправлять update
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}
