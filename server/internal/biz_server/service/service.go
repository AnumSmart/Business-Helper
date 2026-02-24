package service

import (
	"context"
	"fmt"
	pb "global_models/grpc/bot"
	"server/internal/biz_server/repository"
	"server/internal/domain"
	"time"
)

// описание интерфейса сервисного слоя
type BizServiceInterface interface {
	GetEcho() string
}

// MessageServiceInterface - интерфейс для бизнес-логики сообщений
type MessageServiceInterface interface {
	// ProcessMessage - обработка входящего сообщения
	ProcessMessage(ctx context.Context, msg *pb.Message) (*pb.UpdateResponse, error)

	// ProcessCallback - обработка callback от inline клавиатуры
	ProcessCallback(ctx context.Context, callback *pb.CallbackQuery) (*pb.UpdateResponse, error)

	SendMessage(ctx context.Context, req *pb.SendMessageRequest) (*pb.SendMessageResponse, error)
}

// описание структуры сервисного слоя
type BizService struct {
	repo *repository.BizRepository // слой репоизтория (прямая зависимость)
}

// Конструктор возвращает интерфейс
func NewBizService(repo *repository.BizRepository) (BizServiceInterface, error) {
	// проверяем, что на входе интерфейс не nil
	if repo == nil {
		return nil, fmt.Errorf("repo must not be nil")
	}

	return &BizService{
		repo: repo,
	}, nil
}

// метод сервисного слоя для тестирования
func (s *BizService) GetEcho() string {
	return fmt.Sprintf("%s ---> Service layer", s.repo.Echo())
}

// ProcessMessage - обработка входящего сообщения
func (s *BizService) ProcessMessage(ctx context.Context, msg *pb.Message) (*pb.UpdateResponse, error) {
	// 1. Сохраняем входящее сообщение в БД
	incomingMsg := &domain.Message{
		MessageID: msg.MessageId,
		ChatID:    msg.ChatId,
		UserID:    msg.UserId,
		Text:      msg.Text,
		Direction: "incoming",
		Timestamp: time.Unix(msg.Date, 0),
	}

	if err := s.repo.Save(incomingMsg); err != nil {
		return nil, fmt.Errorf("failed to save incoming message: %w", err)
	}

	// 2. Бизнес-логика обработки сообщения
	replyText := s.generateReply(msg.Text, msg.From)

	// 3. Создаем исходящее сообщение и сохраняем в БД
	outgoingMsg := &domain.Message{
		ChatID:    msg.ChatId,
		UserID:    msg.UserId,
		Text:      replyText,
		Direction: "outgoing",
		Timestamp: time.Now(),
	}

	if err := s.repo.Save(outgoingMsg); err != nil {
		return nil, fmt.Errorf("failed to save outgoing message: %w", err)
	}

	// 4. Формируем ответ для бота
	response := &pb.UpdateResponse{
		Success: true,
		Messages: []*pb.OutgoingMessage{
			{
				ChatId: msg.ChatId,
				Text:   replyText,
				// Можно добавить клавиатуру если нужно
				ReplyMarkup: s.createTestKeyboard(),
			},
		},
	}

	return response, nil
}

// generateReply генерирует ответ на сообщение
func (s *BizService) generateReply(text string, user *pb.User) string {
	// Простая логика для примера
	// В реальном проекте здесь может быть AI, бизнес-правила и т.д.

	if text == "/start" {
		return "Добро пожаловать! Я бот-помощник. Чем могу помочь?"
	}

	if text == "/help" {
		return "Доступные команды:\n/start - начало работы\n/help - помощь"
	}

	// Приветствие с именем пользователя
	if user != nil && user.FirstName != "" {
		return fmt.Sprintf("Привет, %s! Вы написали: %s", user.FirstName, text)
	}

	return fmt.Sprintf("Эхо: %s", text)
}

// createTestKeyboard создает тестовую inline клавиатуру
func (s *BizService) createTestKeyboard() *pb.ReplyMarkup {
	return &pb.ReplyMarkup{
		Type: &pb.ReplyMarkup_InlineKeyboard{
			InlineKeyboard: &pb.InlineKeyboardMarkup{
				Rows: []*pb.InlineKeyboardRow{
					{
						Buttons: []*pb.InlineKeyboardButton{
							{
								Text:         "Помощь",
								CallbackData: "help",
							},
						},
					},
					{
						Buttons: []*pb.InlineKeyboardButton{
							{
								Text:         "Еще",
								CallbackData: "more",
							},
							{
								Text: "Сайт",
								Url:  "https://example.com",
							},
						},
					},
				},
			},
		},
	}
}

// ProcessCallback - обработка callback от inline клавиатуры
func (s *BizService) ProcessCallback(ctx context.Context, callback *pb.CallbackQuery) (*pb.UpdateResponse, error) {
	// 1. Логируем callback
	s.repo.SaveCallback(&domain.CallbackLog{
		CallbackID: callback.Id,
		UserID:     callback.UserId,
		ChatID:     callback.ChatId,
		MessageID:  callback.MessageId,
		Data:       callback.Data,
		Timestamp:  time.Now(),
	})

	// 2. Анализируем данные callback и формируем ответ
	response := &pb.UpdateResponse{
		Success: true,
	}

	switch callback.Data {
	case "help":
		// Отправляем новое сообщение с помощью
		response.Messages = append(response.Messages, &pb.OutgoingMessage{
			ChatId: callback.ChatId,
			Text:   "Я бот-помощник. Доступные команды:\n/help - помощь\n/start - начало",
		})

	case "more":
		// Редактируем существующее сообщение (меняем текст)
		// Для редактирования нужно добавить поле в protobuf, пока просто шлем новое
		response.Messages = append(response.Messages, &pb.OutgoingMessage{
			ChatId: callback.ChatId,
			Text:   "Дополнительная информация...",
		})

	default:
		// Ответ на неизвестную команду
		response.Messages = append(response.Messages, &pb.OutgoingMessage{
			ChatId: callback.ChatId,
			Text:   fmt.Sprintf("Неизвестная команда: %s", callback.Data),
		})
	}

	return response, nil
}

func (s BizService) SendMessage(ctx context.Context, req *pb.SendMessageRequest) (*pb.SendMessageResponse, error) {
	// заглушка.........................
	return nil, nil
}
