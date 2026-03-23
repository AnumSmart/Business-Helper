package handlers

import (
	"bot/internal/domain"
	"bot/internal/server/http_server/converter"
	"bot/internal/server/service"
	"context"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	tele "gopkg.in/telebot.v4"
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

// HandleWebhook - основной метод обработки входящих вебхуков от Telegram, режим webhook
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

// хэндлер для обработки команды /start от телеграмм бота в polling режиме
func (h *BotHttpHandler) HandleBotStart(c tele.Context) error {
	args := strings.Fields(c.Text())

	var welcomeMsg string
	var inlineBtns *tele.ReplyMarkup
	var replyMarkup *tele.ReplyMarkup

	// Проверяем наличие параметра
	if len(args) > 1 {
		param := args[1]

		// Обрабатываем разные параметры
		switch param {
		case "menu":
			welcomeMsg = "🏠 Главное меню\n\nДобро пожаловать! Я бот-ассистент. Выберите действие:"
			inlineBtns = &tele.ReplyMarkup{
				InlineKeyboard: [][]tele.InlineButton{
					{
						{Text: "❓ Помощь", Data: "help"},
						{Text: "📚 Ознакомиться", Data: "lookup"},
					},
					{
						{Text: "⚙️ Настройки", Data: "settings"},
					},
				},
			}

		case "help":
			welcomeMsg = "🆘 Справка\n\nЯ могу помочь вам:\n• Ознакомиться с примерами работ\n• Ответить на вопросы\n• Настроить уведомления\n\nВыберите действие:"
			inlineBtns = &tele.ReplyMarkup{
				InlineKeyboard: [][]tele.InlineButton{
					{
						{Text: "📚 Примеры работ", Data: "lookup"},
						{Text: "⚙️ Настройки", Data: "settings"},
					},
					{
						{Text: "🏠 Главное меню", Data: "menu"},
					},
				},
			}

		case "referral":
			welcomeMsg = "🎉 Добро пожаловать по реферальной ссылке!\n\nВам начислен бонус! Ознакомьтесь с нашими возможностями:"
			inlineBtns = &tele.ReplyMarkup{
				InlineKeyboard: [][]tele.InlineButton{
					{
						{Text: "📚 Начать", Data: "lookup"},
						{Text: "❓ Помощь", Data: "help"},
					},
				},
			}
			// Здесь можно сохранить информацию о реферале
			// h.saveReferral(c.Chat().ID, args[2] если есть)

		default:
			// Неизвестный параметр - показываем стандартное приветствие
			welcomeMsg = "Добро пожаловать! Я бот-ассистент. Вы можете ознакомиться с примерами работ. Пожалуйста, выберите одну из функций"
			inlineBtns = &tele.ReplyMarkup{
				InlineKeyboard: [][]tele.InlineButton{
					{
						{Text: "❓ Помощь", Data: "help"},
						{Text: "📚 Ознакомиться", Data: "lookup"},
					},
				},
			}
		}
	} else {
		// Нет параметра - обычный запуск
		welcomeMsg = "Добро пожаловать! Я бот-ассистент. Вы можете ознакомиться с примерами работ. Пожалуйста, выберите одну из функций"
		inlineBtns = &tele.ReplyMarkup{
			InlineKeyboard: [][]tele.InlineButton{
				{
					{Text: "❓ Помощь", Data: "help"},
					{Text: "📚 Ознакомиться", Data: "lookup"},
				},
			},
		}
	}

	// Отправляем сообщение с inline кнопками
	err := c.Send(welcomeMsg, inlineBtns)
	if err != nil {
		return err
	}

	// Reply клавиатура (внизу экрана) - общая для всех случаев
	replyMarkup = &tele.ReplyMarkup{
		ResizeKeyboard: true,
	}

	replyMarkup.Reply(
		replyMarkup.Row(
			replyMarkup.Text("🏠 Главное меню"),
		),
	)

	return c.Send("Используйте кнопки ниже для быстрого доступа:", replyMarkup)
}

// хэндлер для обработки всех текстовых сообщений от телеграмм бота в polling режиме
func (h *BotHttpHandler) HandleBotMessage(c tele.Context) error {

	//конвертируем информацию из контектса сообщения в доменную модель
	update, err := converter.ConvertToUpdate(c)
	if err != nil {
		// Логируем ошибку конвертации
		log.Printf("❌ Ошибка конвертации update: %v", err)
		return c.Send("⚠️ Внутренняя ошибка формата")
	}

	// Создаём стандартный контекст для бизнес-логики
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Конвертируем Telegram формат в gRPC формат
	grpcUpdate := converter.ConvertToGRPCUpdate(update)

	// Отправляем на gRPC сервер для бизнес-логики
	// передает контекст логики в gRPC вызов
	resp, err := h.BotService.ProcessUpdate(ctx, grpcUpdate)
	if err != nil {
		log.Printf("❌ Ошибка gRPC: %v", err)

		// Отправляем пользователю понятное сообщение
		return c.Send("🔌 Сервер временно недоступен. Попробуйте позже.")
	}

	if !resp.Success {
		log.Printf("⚠️ gRPC вернул ошибку: %v", resp.Error)
		return c.Send("⚠️ Не удалось обработать запрос")
	}

	// Если сервер вернул сообщения для отправки - отправляем их в Telegram
	if len(resp.Messages) > 0 {
		// тут вызывается http клиент из сервисного слоя и передаёт ответ боту
		if err := h.BotService.SendHTTPMessages(resp.Messages); err != nil {
			log.Printf("❌ Ошибка отправки через HTTP клиент: %v", err)
			// Не возвращаем ошибку в Telegram, чтобы не было ретраев
			c.Send("⚠️ Сообщение получено, но не доставлено")
			return nil
		}
	}
	return nil
}

// хэндлер для обработки callback-запросов от inline клавиатур от телеграмм бота в polling режиме
func (h *BotHttpHandler) HandleBotCallback(c tele.Context) error {
	log.Printf("📞 Получен callback: data=%s", c.Callback().Data)

	// 1️⃣ Конвертируем
	update, err := converter.ConvertToUpdate(c)
	if err != nil {
		log.Printf("❌ Ошибка конвертации: %v", err)
		c.Respond(&tele.CallbackResponse{
			Text: "❌ Ошибка",
		})
		return err
	}

	// 2️⃣ Создаём контекст с таймаутом
	grpcUpdate := converter.ConvertToGRPCUpdate(update)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// 3️⃣ Отправляем запрос
	resp, err := h.BotService.ProcessUpdate(ctx, grpcUpdate)

	// 4️⃣ Отвечаем на callback (всегда!)
	if err != nil {
		log.Printf("❌ Ошибка gRPC: %v", err)
		return c.Respond(&tele.CallbackResponse{
			Text: "❌ Сервер недоступен",
		})
	}

	if !resp.Success {
		log.Printf("⚠️ Ошибка: %v", resp.Error)
		return c.Respond(&tele.CallbackResponse{
			Text: "⚠️ " + resp.Error,
		})
	}

	log.Println(resp.Messages)

	// 5️⃣ Отправляем сообщения
	if len(resp.Messages) > 0 {
		if err := h.BotService.SendHTTPMessages(resp.Messages); err != nil {
			log.Printf("❌ Ошибка отправки: %v", err)
			return c.Respond(&tele.CallbackResponse{
				Text: "⚠️ Частичный успех",
			})
		}
	}

	// 6️⃣ Успех!
	return c.Respond(&tele.CallbackResponse{
		Text: "✓ Готово!",
	})
}
