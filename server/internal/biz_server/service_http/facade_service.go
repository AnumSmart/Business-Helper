package servicehttp

import "fmt"

// структура для сервисного http слоя
type BizServiceFacade struct{}

// конструктор для сервисного http слоя
func NewBizServiceFacade() *BizServiceFacade {
	return &BizServiceFacade{}
}

func (b *BizServiceFacade) GetEcho() string {
	return fmt.Sprintf("Hello from http service layer")
}
