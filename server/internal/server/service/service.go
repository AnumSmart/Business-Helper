package service

import (
	"fmt"
	"server/internal/server/repository"
)

// описание интерфейса сервисного слоя
type BizServiceInterface interface {
	GetEcho() string
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
