package repository

import (
	"global_models/global_db"
)

// создаём репозиторий базы данных для сервиса авторизации на базе адаптера к pgxpool

// Реализуем ТОЛЬКО то, что нужно auth_service
type bizDBRepository struct {
	Pool global_db.Pool // строится на базе глобального интерфейса
}

// создаём конструктор для слоя базы данных
func NewBizDBRepository(pool global_db.Pool) *bizDBRepository {
	return &bizDBRepository{Pool: pool}
}
