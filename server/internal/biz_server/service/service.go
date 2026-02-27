package service

import (
	"context"
	"fmt"
	"server/internal/biz_server/grpcclient"
	"server/internal/biz_server/repository"
	"server/internal/domain"
)

// описание интерфейса сервисного слоя для HTTP сервера
type ServiceForHTTPHandler interface {
	GetEcho() string
}

// ServiceForGRPCHandler - интерфейс для бизнес-логики сообщений (отвечаем, если есть запросы со стороны бота), работа с GRPC
type ServiceForGRPCHandler interface {
	// проверяем и сохраняем сообщение
	CheckAndSaveMsg(msg *domain.Message) error
	CheckAndSaveCallBack(callBackLog *domain.CallbackLog) error
	// генерируем ответ
	GenerateReply(text string, user *domain.User) string
	CreateTestKeyboard() *domain.ReplyMarkup
	AnswerIncomingMsg(ctx context.Context, req *domain.IncomingMessage) (*domain.MessageResponse, error)
}

// описание структуры сервисного слоя
type BizService struct {
	repo       *repository.BizRepository // слой репоизтория (прямая зависимость)
	grpcClient *grpcclient.BotGrpcClient // grpc клиент
}

// Конструктор возвращает интерфейс
func NewBizService(repo *repository.BizRepository, grpcClient *grpcclient.BotGrpcClient) (*BizService, error) {
	// проверяем, что на входе интерфейс не nil
	if repo == nil {
		return nil, fmt.Errorf("repo must not be nil")
	}

	if grpcClient == nil {
		return nil, fmt.Errorf("grpcClient must not be nil")
	}

	return &BizService{
		repo:       repo,
		grpcClient: grpcClient,
	}, nil
}

// метод сервисного слоя для тестирования
func (s *BizService) GetEcho() string {
	return fmt.Sprintf("%s ---> Service layer", s.repo.Echo())
}

// метод для проверки и сохданения входящего сообщения (по GRPC) в базу
func (s *BizService) CheckAndSaveMsg(msg *domain.Message) error {
	// возможные проверки.....
	if msg == nil {
		return fmt.Errorf("Incoming messgage can not be nil! [error in service layer]")
	}

	// сохраняем входящее сообщение в базу данных
	if err := s.repo.Save(msg); err != nil {
		return fmt.Errorf("failed to save outgoing message: %w", err)
	}

	return nil
}

// generateReply генерирует ответ на сообщение
func (s *BizService) GenerateReply(text string, user *domain.User) string {
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

// CreateTestReplyKeyboard создает тестовую обычную клавиатуру
// (альтернативный пример для полноты)
func (s *BizService) CreateTestKeyboard() *domain.ReplyMarkup {
	return &domain.ReplyMarkup{
		Keyboard: [][]domain.Button{
			// Первый ряд
			{
				{Text: "Кнопка 1"},
				{Text: "Кнопка 2"},
			},
			// Второй ряд
			{
				{Text: "Отмена"},
			},
		},
		ResizeKeyboard:  true, // Подогнать размер под кнопки
		OneTimeKeyboard: true, // Спрятать после использования
	}
}

// метод для проверки и сохданения callback в базу
func (s *BizService) CheckAndSaveCallBack(callBackLog *domain.CallbackLog) error {
	if callBackLog == nil {
		return fmt.Errorf("callBackLog can not be nil! [error in service layer]")
	}

	// сохраняем callBack в базу
	err := s.repo.SaveCallback(callBackLog)
	if err != nil {
		return err
	}
	return nil
}

// метод для обработки сообщения от grpc клиента и ответа
func (s *BizService) AnswerIncomingMsg(ctx context.Context, req *domain.IncomingMessage) (*domain.MessageResponse, error) {
	// log.Printf("Send message to chat %d: %s", req.ChatId, req.Text)

	// Здесь может быть валидация, сохранение в БД, etc.
	return &domain.MessageResponse{
		Success: true,
	}, nil
}
