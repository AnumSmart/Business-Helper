package handlers

import (
	"fmt"
	"net/http"
	"server/internal/biz_server/service"

	"github.com/gin-gonic/gin"
)

// описание интерфейса слоя хэндлеров
type BizHandlerInterface interface {
	EchoAuthServer(c *gin.Context) // ЭХО для тестирования!
}

// структура хэндлера сервера авторизации
type BizHandler struct {
	service service.BizServiceInterface
}

// конструктор для слоя хэндлеров
func NewBizHandler(service service.BizServiceInterface) BizHandlerInterface {
	return &BizHandler{
		service: service,
	}
}

// тестовый метод слоя хэндлеров
func (h *BizHandler) EchoAuthServer(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": fmt.Sprintf("Handler layer: %s", h.service.GetEcho())})
}
