package handlers

import (
	grpcclient "bot/internal/server/grpc_client"
	httpclient "bot/internal/server/http_client"
	"net/http"

	"github.com/gin-gonic/gin"

	pb "global_models/grpc/bot"
)

// BotHandler обрабатывает входящие вебхуки от Telegram
// Содержит все необходимые клиенты для обработки запроса
type BotHandler struct {
	GrpcClient *grpcclient.BotGrpcClient // Для отправки данных в gRPC сервер
	TgClient   *httpclient.BotHTTPClient // Для отправки ответов в Telegram
}

// NewBotHandler создает новый экземпляр обработчика с внедренными зависимостями
// Паттерн "Dependency Injection" - клиенты передаются извне
func NewBotHandler(grpcClient *grpcclient.BotGrpcClient, tgClient *httpclient.BotHTTPClient) *BotHandler {
	return &BotHandler{
		GrpcClient: grpcClient,
		TgClient:   tgClient,
	}
}

// TelegramUpdate представляет структуру входящего обновления от Telegram API
// Теги json используются для маппинга полей из JSON в структуру
type TelegramUpdate struct {
	UpdateID int64 `json:"update_id"` // Уникальный ID обновления

	// Message - указатель, т.к. может отсутствовать (если это callback)
	Message *struct {
		MessageID int64 `json:"message_id"` // ID сообщения в чате
		From      struct {
			ID       int64  `json:"id"`       // ID отправителя
			Username string `json:"username"` // Username (без @)
		} `json:"from"`
		Chat struct {
			ID int64 `json:"id"` // ID чата
		} `json:"chat"`
		Date int64  `json:"date"` // Unix timestamp
		Text string `json:"text"` // Текст сообщения
	} `json:"message"`

	// CallbackQuery - указатель, т.к. может отсутствовать (если это сообщение)
	CallbackQuery *struct {
		ID   string `json:"id"` // Уникальный ID callback
		From struct {
			ID int64 `json:"id"` // ID пользователя, нажавшего кнопку
		} `json:"from"`
		Message struct {
			MessageID int64 `json:"message_id"` // ID сообщения с клавиатурой
			Chat      struct {
				ID int64 `json:"id"` // ID чата
			} `json:"chat"`
		} `json:"message"`
		Data string `json:"data"` // Данные из callback_data
	} `json:"callback_query"`
}

// HandleWebhook - основной метод обработки входящих вебхуков от Telegram
// Принимает gin.Context для доступа к запросу и ответу
func (h *BotHandler) HandleWebhook(c *gin.Context) {
	var update TelegramUpdate

	// ShouldBindJSON автоматически парсит JSON из тела запроса в структуру
	// Возвращает ошибку, если JSON невалиден или не соответствует структуре
	if err := c.ShouldBindJSON(&update); err != nil {
		// В случае ошибки возвращаем 400 Bad Request
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Шаг 1: Конвертируем Telegram формат в gRPC формат
	grpcUpdate := convertToGRPCUpdate(&update)

	// Шаг 2: Отправляем на gRPC сервер для бизнес-логики
	// c.Request.Context() передает контекст HTTP запроса в gRPC вызов
	resp, err := h.GrpcClient.ProcessUpdate(c.Request.Context(), grpcUpdate)
	if err != nil {
		// Ошибка связи с gRPC сервером
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Шаг 3: Если сервер вернул сообщения для отправки - отправляем их в Telegram
	if resp.Success && len(resp.Messages) > 0 {
		if err := h.TgClient.SendOutgoingMessages(resp.Messages); err != nil {
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

// convertToGRPCUpdate конвертирует Telegram структуру в protobuf структуру
// Это "адаптер" между внешним API (Telegram) и внутренним (gRPC)
func convertToGRPCUpdate(update *TelegramUpdate) *pb.UpdateRequest {
	// Создаем базовый запрос с update_id
	req := &pb.UpdateRequest{
		UpdateId: update.UpdateID,
	}

	// Если есть сообщение - заполняем структуру Message
	if update.Message != nil {
		req.Message = &pb.Message{
			MessageId: update.Message.MessageID,
			ChatId:    update.Message.Chat.ID,
			UserId:    update.Message.From.ID,
			Text:      update.Message.Text,
			Date:      update.Message.Date,
			From: &pb.User{
				Id:       update.Message.From.ID,
				Username: update.Message.From.Username,
			},
			Chat: &pb.Chat{
				Id:   update.Message.Chat.ID,
				Type: "private", // Упрощение: в реальном проекте нужно определять тип
			},
		}
	}

	// Если есть callback query - заполняем структуру CallbackQuery
	if update.CallbackQuery != nil {
		req.CallbackQuery = &pb.CallbackQuery{
			Id:        update.CallbackQuery.ID,
			UserId:    update.CallbackQuery.From.ID,
			MessageId: update.CallbackQuery.Message.MessageID,
			ChatId:    update.CallbackQuery.Message.Chat.ID,
			Data:      update.CallbackQuery.Data,
		}
	}

	return req
}
