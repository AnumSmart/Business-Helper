package configs

import "pkg/configs"

// структрура конфига для всего сервиса работы с ботами
type BizServiceConfig struct {
	ServerConf     *configs.ServerConfig
	PostgresDBConf *configs.PostgresDBConfig
	RedisConf      *configs.RedisConfig
}

// путь к .env файлу
const (
	envPath = "c:\\Users\\aliaksei.makarevich\\go\\bizhelper v_1_20\\server\\.env"
)
