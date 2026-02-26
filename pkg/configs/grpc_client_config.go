package configs

import (
	"fmt"
	"time"
)

// конфиг для grpc клиента
type GRPCClientConfig struct {
	Host       string        `yaml:"host"`       // хост сервера, к которому подключается клиент
	Port       string        `yaml:"port"`       // порт сервра
	TimeOut    time.Duration `yaml:"timeout"`    // Таймаут для запросов
	MaxRetries int           `yml:"max_retries"` // Количество повторных попыток (опционально)
}

// метод получения адреса
func (c *GRPCClientConfig) Addr() string {
	return fmt.Sprintf("%s:%s", c.Host, c.Port)
}

// дэфолтный конфиг
func UseDefaultGRPCClientConfig() *GRPCClientConfig {
	return &GRPCClientConfig{
		Host:       "localhost",
		Port:       "50050",
		TimeOut:    30 * time.Second,
		MaxRetries: 3,
	}
}
