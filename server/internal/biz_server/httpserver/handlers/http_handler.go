package handlers

import (
	"fmt"
	"net/http"
	"server/internal/biz_server/service"

	"github.com/gin-gonic/gin"
)

// описание интерфейса слоя хэндлеров
type BizHTTPHandlerInterface interface {
	EchoAuthServer(c *gin.Context) // ЭХО для тестирования!
}

// структура хэндлера http сервера основной логики
type BizHTTPHandler struct {
	Service service.BizServiceInterface
}

// конструктор для слоя хэндлеров
func NewBizHandler(service service.BizServiceInterface) BizHTTPHandlerInterface {
	return &BizHTTPHandler{
		Service: service,
	}
}

// тестовый метод слоя хэндлеров
func (h *BizHTTPHandler) EchoAuthServer(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": fmt.Sprintf("Handler layer: %s", h.Service.GetEcho())})
}
