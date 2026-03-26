package servicegrpc

import (
	"context"
	"errors"
	"fmt"
	"server/internal/biz_server/repository"
	"server/internal/domain"
	"time"
)

// ========== User Service ==========
type UserService interface {
	RegisterOrUpdate(ctx context.Context, telegramID int64, firstName, lastName, username string) (*domain.User, error)
	GetByTelegramID(ctx context.Context, telegramID int64) (*domain.User, error)
	UpdateActivity(ctx context.Context, telegramID int64) error
}

// структура сервиса пользователей
type userService struct {
	repo *repository.BizRepository
}

// конструктор для сервиса пользователей
func NewUserService(repo *repository.BizRepository) UserService {
	return &userService{repo: repo}
}

// метод для регистрации или обновления текущего пользователя
func (s *userService) RegisterOrUpdate(ctx context.Context, telegramID int64, firstName, lastName, username string) (*domain.User, error) {
	user, err := s.repo.GetUserByTelegramID(ctx, telegramID)
	if err != nil && !errors.Is(err, repository.ErrUserNotFound) {
		return nil, fmt.Errorf("failed to check user existence: %w", err)
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

		return user, nil
	}

	// Обновление существующего пользователя
	return s.updateExistingUser(ctx, user, firstName, lastName, username, now)
}

// Обновление существующего пользователя
func (s *userService) updateExistingUser(ctx context.Context, user *domain.User, firstName, lastName, username string, now time.Time) (*domain.User, error) {
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
		if err := s.repo.UpdateLastSeen(ctx, user.TelegramID); err != nil {
			return nil, fmt.Errorf("failed to update last_seen: %w", err)
		}
	}

	return user, nil
}

// GetByTelegramID - получение пользователя
func (s *userService) GetByTelegramID(ctx context.Context, telegramID int64) (*domain.User, error) {
	user, err := s.repo.GetUserByTelegramID(ctx, telegramID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	return user, nil
}

// UpdateActivity - быстрое обновление только времени активности
func (s *userService) UpdateActivity(ctx context.Context, telegramID int64) error {
	if err := s.repo.UpdateLastSeen(ctx, telegramID); err != nil {
		return fmt.Errorf("failed to update activity: %w", err)
	}
	return nil
}
