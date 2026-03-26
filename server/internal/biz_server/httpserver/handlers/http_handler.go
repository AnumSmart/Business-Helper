package handlers

import (
	"fmt"
	"global_models/interf"
	"net/http"
	servicehttp "server/internal/biz_server/service_http"

	"github.com/gin-gonic/gin"
)

// структура хэндлера http сервера основной логики
type BizHTTPHandler struct {
	Service *servicehttp.BizServiceFacade
}

// конструктор для слоя хэндлеров
func NewBizHandler(service *servicehttp.BizServiceFacade) interf.BizHTTPHandlerInterface {
	return &BizHTTPHandler{
		Service: service,
	}
}

// тестовый метод слоя хэндлеров
func (h *BizHTTPHandler) EchoServer(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": fmt.Sprintf("Handler layer: %s", h.Service.GetEcho())})
}
