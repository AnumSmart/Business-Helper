package repository

import (
	"fmt"
	"global_models/global_cache"
	"server/internal/interfaces"
)

// создаём репозиторий кэша (тут редис) на базе глобального интерфейса

// Реализуем ТОЛЬКО то, что нужно слоя service
type bizCacheRepository struct {
	blackCache global_cache.Cache // создаём на базе глобального интерфейса
	prefix     string
}

// конструктор для репозитория черного списка (использует интерфейс для глобального кэша)
func NewBizCacheRepo(cache global_cache.Cache, prefix string) (interfaces.CacheRepoInterface, error) {
	// Проверяем обязательные зависимости
	if cache == nil {
		return nil, fmt.Errorf("cache cannot be nil")
	}

	// Проверяем префикс
	if prefix == "" {
		return nil, fmt.Errorf("prefix cannot be empty")
	}
	return &bizCacheRepository{
		blackCache: cache,
		prefix:     prefix,
	}, nil
}
