package handlers

import (
	"fmt"
	"global_models/interf"
	"net/http"
	"server/internal/biz_server/service"

	"github.com/gin-gonic/gin"
)

// структура хэндлера http сервера основной логики
type BizHTTPHandler struct {
	Service service.ServiceForHTTPHandler
}

// конструктор для слоя хэндлеров
func NewBizHandler(service service.ServiceForHTTPHandler) interf.BizHTTPHandlerInterface {
	return &BizHTTPHandler{
		Service: service,
	}
}

// тестовый метод слоя хэндлеров
func (h *BizHTTPHandler) EchoServer(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": fmt.Sprintf("Handler layer: %s", h.Service.GetEcho())})
}
