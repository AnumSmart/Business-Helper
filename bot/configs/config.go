package configs

import (
	"fmt"
	"os"
	"pkg/configs"

	"github.com/joho/godotenv"
)

// структрура конфига для сервиса сервера бота
type BotServiceConfig struct {
	ServerConf *configs.ServerConfig
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

	// загружаем данные из .yml файла для serverConfig
	serverConfig, err := configs.LoadYAMLConfig[configs.ServerConfig](os.Getenv("BOT_SERVER_CONFIG_ADDRESS_STRING"), configs.UseDefaultServerConfig)
	if err != nil {
		return nil, fmt.Errorf("Error during loading config: %s\n", err.Error())
	}

	return &BotServiceConfig{
		ServerConf: serverConfig,
	}, nil
}
