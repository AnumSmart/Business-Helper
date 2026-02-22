package config

import (
	"fmt"
	"os"
	"pkg/configs"

	"github.com/joho/godotenv"
)

type BotConfig struct {
	BotToken    string `yaml:"bot_token"`    // BotToken - это уникальный идентификатор бота в Telegram (Выдается @BotFather при создании бота)
	WebhookURL  string `yaml:"webhook_url"`  // Публичный HTTPS URL, на который Telegram будет отправлять обновления
	WebhookPort string `yaml:"webhook_port"` // Локальный порт, на котором бот слушает входящие вебхуки. Обычно 8080, 8443 или 443 (для HTTPS)
	GRPCServer  string `yaml:"grpc_server"`  // адрес gRPC сервера, к которому подключается бот, Бот выступает как gRPC клиент и шлет сюда обновления от Telegram
	// Стандартные значения:
	//   - "development" - локальная разработка (больше логов, debug режим)
	//   - "staging" - тестовый сервер (похоже на production, но с тестовыми данными)
	//   - "production" - продакшен (минимум логов, максимум производительности)
	Environment string `yaml:"environment"`
}

const (
	envPath = "c:\\Users\\aliaksei.makarevich\\go\\bizhelper_v_1_20\\bot\\.env"
)

func LoadConfig() (config *BotConfig, err error) {
	err = godotenv.Load(envPath)
	if err != nil {
		return nil, fmt.Errorf("Error during loading config: %s\n", err.Error())
	}

	// загружаем конфиг из .yml файла
	botConfig, err := configs.LoadYAMLConfig[BotConfig](os.Getenv("BOT_CONFIG_ADDRESS_STRING"), UseDefaultBotConfig)
	if err != nil {
		return nil, fmt.Errorf("Error during loading config: %s\n", err.Error())
	}

	return botConfig, nil
}

func UseDefaultBotConfig() *BotConfig {
	return &BotConfig{
		BotToken:    "123456:ABC",
		WebhookURL:  "https://example.com/webhook",
		WebhookPort: "8080",
		GRPCServer:  "localhost:50051",
		Environment: "production",
	}
}
