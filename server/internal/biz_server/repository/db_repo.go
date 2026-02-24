package repository

import (
	"global_models/global_db"
	"server/internal/interfaces"
)

// создаём репозиторий базы данных для сервиса авторизации на базе адаптера к pgxpool

// Реализуем ТОЛЬКО то, что нужно auth_service
type bizDBRepository struct {
	pool global_db.Pool // строится на базе глобального интерфейса
}

// создаём конструктор для слоя базы данных
func NewBizDBRepository(pool global_db.Pool) interfaces.DBRepoInterface {
	return &bizDBRepository{pool: pool}
}
