package configs

import (
	"fmt"
	"time"
)

// структура конфига grpc сервера
type GRPCServerConfig struct {
	Host                  string        `yaml:"host"`
	Port                  string        `yaml:"port"`
	MaxConnectionIdle     time.Duration `yaml:"max_connection_idle"`
	MaxConnectionAge      time.Duration `yaml:"max_connection_age"`
	MaxConnectionAgeGrace time.Duration `yaml:"max_connection_age_grace"`
	KeepaliveTime         time.Duration `yaml:"keepalive_time"`
	KeepaliveTimeout      time.Duration `yaml:"keepalive_timeout"`
	MaxRecvMsgSize        int           `yaml:"max_recv_msg_size"`
	MaxSendMsgSize        int           `yaml:"max_send_msg_size"`
}

// метод получения адреса
func (c *GRPCServerConfig) Addr() string {
	return fmt.Sprintf("%s:%s", c.Host, c.Port)
}

// дэфолтный конфиг
func UseDefaultGRPCServerConfig() *GRPCServerConfig {
	return &GRPCServerConfig{
		Host:                  "0.0.0.0",
		Port:                  "50051",
		MaxConnectionIdle:     15 * time.Minute,
		MaxConnectionAge:      30 * time.Minute,
		MaxConnectionAgeGrace: 5 * time.Minute,
		KeepaliveTime:         5 * time.Minute,
		KeepaliveTimeout:      20 * time.Second,
		MaxRecvMsgSize:        10485760,
		MaxSendMsgSize:        10485760,
	}
}
