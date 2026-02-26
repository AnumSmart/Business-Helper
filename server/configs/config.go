package configs

import (
	"fmt"
	"os"
	"pkg/configs"

	"github.com/joho/godotenv"
)

// структрура конфига для всего сервиса работы с ботами
type BizServiceConfig struct {
	HTTPServerConf   *configs.ServerConfig     // конфиг для HTTP сервера
	GRPCServerConf   *configs.GRPCServerConfig // конфиг для GRPC сервера
	GRPCClientConfig *configs.GRPCClientConfig // конфиг для GRPC клиента
	PostgresDBConf   *configs.PostgresDBConfig // конфиг для базы данных POSTGRES
	RedisConf        *configs.RedisConfig      // конфиг для кэша REDIS
}

// путь к .env файлу
const (
	envPath = "c:\\Son_Alex\\GO_projects\\biz_helper\\server\\.env"
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

	// загружаем конфиг для grpc сервера
	grpcServerConfig, err := configs.LoadYAMLConfig[configs.GRPCServerConfig](os.Getenv("GRPC_SERVER_CONFIG_ADDRESS_STRING"), configs.UseDefaultGRPCServerConfig)
	if err != nil {
		return nil, fmt.Errorf("Error during loading config: %s\n", err.Error())
	}

	// загружаем конфиг для grpc клиента
	grpcClientCofig, err := configs.LoadYAMLConfig[configs.GRPCClientConfig](os.Getenv("GRPC_CLIENT_CONFIG_ADDRESS_STRING"), configs.UseDefaultGRPCClientConfig)

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
		HTTPServerConf:   serverConfig,
		GRPCServerConf:   grpcServerConfig,
		GRPCClientConfig: grpcClientCofig,
		PostgresDBConf:   postgresDBConfig,
		RedisConf:        redisConfig,
	}, nil
}
