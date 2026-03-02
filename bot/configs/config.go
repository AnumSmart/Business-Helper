package configs

import (
	"fmt"
	"os"
	"pkg/configs"

	"github.com/joho/godotenv"
)

// структрура конфига для сервиса сервера бота
type BotServiceConfig struct {
	HTTPServerConfig *configs.HttpServerConfig
	GRPCServerConfig *configs.GRPCServerConfig
}

// путь к .env файлу
const (
	envPath = "c:\\Son_Alex\\GO_projects\\biz_helper\\server\\.env"
)

// загружаем конфиг-данные из .env
func LoadBotServiceConfig() (*BotServiceConfig, error) {
	err := godotenv.Load(envPath)
	if err != nil {
		return nil, fmt.Errorf("Error during loading config: %s\n", err.Error())
	}

	// загружаем данные из .yml файла для httpServerConfig
	httpServerConfig, err := configs.LoadYAMLConfig[configs.HttpServerConfig](os.Getenv("BOT_HTTP_SERVER_CONFIG_ADDRESS_STRING"), configs.UseDefaultServerConfig)
	if err != nil {
		return nil, fmt.Errorf("Error during loading config: %s\n", err.Error())
	}

	// загружаем данные из .yml файла для grpcServerConfig
	grpcServerConfig, err := configs.LoadYAMLConfig[configs.GRPCServerConfig](os.Getenv("BOT_GRPC_SERVER_CONFIG_ADDRESS_STRING"), configs.UseDefaultGRPCServerConfig)
	if err != nil {
		return nil, fmt.Errorf("Error during loading config: %s\n", err.Error())
	}

	return &BotServiceConfig{
		HTTPServerConfig: httpServerConfig,
		GRPCServerConfig: grpcServerConfig,
	}, nil
}
