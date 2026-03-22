package service

import (
	"context"
	"errors"
	"fmt"
	"server/internal/biz_server/grpcclient"
	"server/internal/biz_server/repository"
	"server/internal/domain"
	"time"
)

// описание интерфейса сервисного слоя для HTTP сервера
type ServiceForHTTPHandler interface {
	GetEcho() string
}

// ServiceForGRPCHandler - интерфейс для бизнес-логики сообщений (отвечаем, если есть запросы со стороны бота), работа с GRPC
type ServiceForGRPCHandler interface {
	// проверяем и сохраняем сообщение
	CheckAndSaveMsg(cxt context.Context, msg *domain.Message) error
	CheckAndSaveCallBack(cxt context.Context, callBackLog *domain.CallbackLog) error
	// генерируем ответ
	GenerateReply(text string, user *domain.User) string
	CreateTestKeyboard() *domain.ReplyMarkup
	CreateWelcomeReplyKeyboard() *domain.ReplyMarkup
	AnswerIncomingMsg(ctx context.Context, req *domain.IncomingMessage) (*domain.MessageResponse, error)
	// RegisterOrUpdate - регистрирует нового или обновляет существующего пользователя
	// Вызывается при каждом обращении к боту
	RegisterOrUpdate(ctx context.Context, telegramID int64, firstName, lastName, username string) (*domain.User, error)
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
func (s *BizService) CheckAndSaveMsg(ctx context.Context, msg *domain.Message) error {
	// возможные проверки.....
	if msg == nil {
		return fmt.Errorf("Incoming messgage can not be nil! [error in service layer]")
	}

	// сохраняем входящее сообщение в базу данных
	if err := s.repo.Save(ctx, msg); err != nil {
		return fmt.Errorf("failed to save outgoing message: %w", err)
	}

	return nil
}

// generateReply генерирует ответ на сообщение
func (s *BizService) GenerateReply(text string, user *domain.User) string {
	// Простая логика для примера
	// В реальном проекте здесь может быть AI, бизнес-правила и т.д.

	if text == "/start" {
		return "Добро пожаловать! Я бот-помощник. Если вы хотите посмотреть работы по организации дизайна - нажмите кнопку 'Посмотреть'"
	}

	if text == "/help" {
		return "Этот бот создан, чтобы облегчить вам жизнь"
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

// CreateWelcomeReplyKeyboard создает приветственную клавиатуру
func (s *BizService) CreateWelcomeReplyKeyboard() *domain.ReplyMarkup {
	return &domain.ReplyMarkup{
		InlineKeyboard: [][]domain.InlineButton{
			// Первый ряд
			{
				{Text: "Поисковик",
					URL: "https://www.google.com/"},
			},
		},
		ResizeKeyboard:  true, // Подогнать размер под кнопки
		OneTimeKeyboard: true, // Спрятать после использования
	}
}

// метод для проверки и сохданения callback в базу
func (s *BizService) CheckAndSaveCallBack(ctx context.Context, callBackLog *domain.CallbackLog) error {
	if callBackLog == nil {
		return fmt.Errorf("callBackLog can not be nil! [error in service layer]")
	}

	// сохраняем callBack в базу
	err := s.repo.SaveCallback(ctx, callBackLog)
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

// метод для регистрации или обновления текущего пользователя
func (s *BizService) RegisterOrUpdate(ctx context.Context, telegramID int64, firstName, lastName, username string) (*domain.User, error) {
	// Проверяем, существует ли пользователь

	user, err := s.repo.GetUserByTelegramID(ctx, telegramID)
	if err != nil {
		// Если пользователь не найден - это нормально, просто создадим нового
		if errors.Is(err, repository.ErrUserNotFound) {
			user = nil
		} else {
			// Другая ошибка (проблемы с БД и т.д.)
			return nil, fmt.Errorf("failed to check user existence: %w", err)
		}
	}

	now := time.Now()

	if user == nil {
		// Создаём нового пользователя
		user = &domain.User{
			TelegramID: telegramID,
			FirstName:  firstName,
			LastName:   lastName,
			Username:   username,
			IsActive:   true,
			CreatedAt:  now,
			LastSeenAt: now,
		}

		if err := s.repo.CreateUser(ctx, user); err != nil {
			return nil, fmt.Errorf("failed to create user: %w", err)
		}

		// Можно залогировать нового пользователя
		fmt.Printf("🎉 New user registered: %d (@%s)\n", telegramID, username)
	} else {
		// Обновляем только если данные изменились
		needsUpdate := false

		if user.FirstName != firstName {
			user.FirstName = firstName
			needsUpdate = true
		}
		if user.LastName != lastName {
			user.LastName = lastName
			needsUpdate = true
		}
		if user.Username != username {
			user.Username = username
			needsUpdate = true
		}

		// Всегда обновляем время активности
		user.LastSeenAt = now

		if needsUpdate {
			if err := s.repo.Update(ctx, user); err != nil {
				return nil, fmt.Errorf("failed to update user: %w", err)
			}
		} else {
			// Только обновляем last_seen
			if err := s.repo.UpdateLastSeen(ctx, telegramID); err != nil {
				return nil, fmt.Errorf("failed to update last_seen: %w", err)
			}
		}
	}

	return user, nil
}

// GetByTelegramID - получение пользователя
func (s *BizService) GetByTelegramID(ctx context.Context, telegramID int64) (*domain.User, error) {
	user, err := s.repo.GetUserByTelegramID(ctx, telegramID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	return user, nil
}

// UpdateActivity - быстрое обновление только времени активности
func (s *BizService) UpdateActivity(ctx context.Context, telegramID int64) error {
	if err := s.repo.UpdateLastSeen(ctx, telegramID); err != nil {
		return fmt.Errorf("failed to update activity: %w", err)
	}
	return nil
}
