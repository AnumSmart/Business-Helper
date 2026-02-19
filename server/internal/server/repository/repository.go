package repository

import (
	"fmt"
	"server/internal/interfaces"
)

// описание структуры слоя репозитория
type BizRepository struct {
	DBRepo    interfaces.DBRepoInterface
	CacheRepo interfaces.CacheRepoInterface
}

// конструктор для слоя репозиторий
func NewBizRepository(dbRepo interfaces.DBRepoInterface, cacheRepo interfaces.CacheRepoInterface) (*BizRepository, error) {
	// Проверяем обязательные зависимости
	if dbRepo == nil {
		return nil, fmt.Errorf("dbRepo is required")
	}
	if cacheRepo == nil {
		return nil, fmt.Errorf("blackListRepo is required")
	}
	return &BizRepository{
		DBRepo:    dbRepo,
		CacheRepo: cacheRepo,
	}, nil
}

// метод для теста
func (r *BizRepository) Echo() string {
	return fmt.Sprintln("Hello from repo layer!")
}
