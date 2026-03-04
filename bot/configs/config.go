package configs

import (
	"bot/internal/config"
	"fmt"
	"os"
	"pkg/configs"

	"github.com/joho/godotenv"
)

// структрура конфига для сервиса сервера бота
type BotServiceConfig struct {
	HTTPServerConfig *config.BotHttpServerConfig
	GRPCServerConfig *configs.GRPCServerConfig
}

// путь к .env файлу
const (
	envPath = "c:\\Son_Alex\\GO_projects\\biz_helper\\bot\\.env"
)

// загружаем конфиг-данные из .env
func LoadBotServiceConfig() (*BotServiceConfig, error) {
	err := godotenv.Load(envPath)
	if err != nil {
		return nil, fmt.Errorf("Error during loading config: %s\n", err.Error())
	}

	// загружаем данные конфига для HTTP сервера
	httpServerConfig, err := config.LoadBotHttpServerConfig()
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
