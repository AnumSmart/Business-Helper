package main

import (
	"bot/internal/dependencies"
	"bot/internal/server"
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	// Создаем корневой контекст
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Инициализируем общие зависимости
	deps, err := dependencies.InitDependencies(ctx)
	if err != nil {
		log.Fatalf("Failed to initialize dependencies: %v", err)
	}

	// Создаем HTTP-сервер
	server, err := server.NewBotServiceServer(ctx, deps.BotServerconfig.ServerConf, deps.BotHandler)
	if err != nil {
		panic("Failed to create server!")
	}

	// создаём канал, который бдут реагировать на системные сигналы
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Запуск сервера
	go func() {
		fmt.Printf("🚀 HTTP сервер бота запускается на %s\n", deps.BotServerconfig.ServerConf.Addr())
		if err := server.Run(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed: %v", err)
		}
	}()

	// Ожидание сигнала
	<-sigChan
	fmt.Println("\n🛑 Остановка сервера бота...")

	// Graceful shutdown
	shutdownCtx, shutdownCancel := context.WithTimeout(ctx, 30*time.Second)
	defer shutdownCancel()

	// Останавливаем HTTP сервер (ждем текущие запросы)
	fmt.Println("Останавливаем HTTP сервер бота и все клиенты внутри этого сервера...")
	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Printf("Error during server shutdown: %v", err)
	}

}
