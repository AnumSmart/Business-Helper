package service

import (
	"context"
	"fmt"
	pb "global_models/grpc/bot"
	"server/internal/biz_server/repository"
	"server/internal/domain"
)

// описание интерфейса сервисного слоя для HTTP сервера
type BizServiceInterface interface {
	GetEcho() string
}

// InMessageServiceInterface - интерфейс для бизнес-логики сообщений (отвечаем, если есть запросы со стороны бота), работа с GRPC
type InMessageServiceInterface interface {
	// проверяем и сохраняем сообщение
	CheckAndSaveMsg(msg *domain.Message) error
	CheckAndSaveCallBack(callBackLog *domain.CallbackLog) error
	// генерируем ответ
	GenerateReply(text string, user *pb.User) string
	CreateTestKeyboard() *pb.ReplyMarkup
}

// OutMessageServiceInterface - интерфейс для бизнес-логики сообщений (отвечаем, боту в зависимости от бизнесс-логики, без его запроса)
// возможно, cron операции
type OutMessageServiceInterface interface {
	SendMessage(ctx context.Context, req *pb.SendMessageRequest) (*pb.SendMessageResponse, error)
}

// описание структуры сервисного слоя
type BizService struct {
	repo *repository.BizRepository // слой репоизтория (прямая зависимость)
}

// Конструктор возвращает интерфейс
func NewBizService(repo *repository.BizRepository) (*BizService, error) {
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

// метод для проверки и сохданения входящего сообщения в базу
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
func (s *BizService) GenerateReply(text string, user *pb.User) string {
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
func (s *BizService) CreateTestKeyboard() *pb.ReplyMarkup {
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

// метод для оправки сообщения от бота без запроса от пользователя (согласно бизнесс-логике)
func (s BizService) SendMessage(ctx context.Context, req *pb.SendMessageRequest) (*pb.SendMessageResponse, error) {
	// log.Printf("Send message to chat %d: %s", req.ChatId, req.Text)

	// Здесь может быть валидация, сохранение в БД, etc.
	return &pb.SendMessageResponse{
		Success: true,
	}, nil
}
