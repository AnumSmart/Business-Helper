package configs

import (
	"fmt"
	"os"
	"pkg/configs"

	"github.com/joho/godotenv"
)

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

// загружаем конфиг-данные из .env
func LoadConfig() (*BizServiceConfig, error) {
	err := godotenv.Load(envPath)
	if err != nil {
		return nil, fmt.Errorf("Error during loading config: %s\n", err.Error())
	}

	// загружаем данные из .yml файла для serverConfig
	serverConfig, err := configs.LoadYAMLConfig[configs.ServerConfig](os.Getenv("SERVER_CONFIG_ADDRESS_STRING"), configs.UseDefaultServerConfig)
	if err != nil {
		return nil, fmt.Errorf("Error during loading config: %s\n", err.Error())
	}

	// загружаем данные из .env файла для postgresDBConfig
	postgresDBConfig, err := configs.NewPostgresDBConfigFromEnv()
	if err != nil {
		return nil, fmt.Errorf("Error during loading config: %s\n", err.Error())
	}

	// загружаем данные из .env файла для redisConfig
	redisConfig, err := configs.NewRedisConfigFromEnv()
	if err != nil {
		return nil, fmt.Errorf("Error during loading config: %s\n", err.Error())
	}

	return &BizServiceConfig{
		ServerConf:     serverConfig,
		PostgresDBConf: postgresDBConfig,
		RedisConf:      redisConfig,
	}, nil
}
