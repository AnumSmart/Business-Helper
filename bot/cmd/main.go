package main

import (
	"bot/internal/dependencies"
	grpcserver "bot/internal/server/grpc_server"
	httpserver "bot/internal/server/http_server"
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

	// Создаем HTTP-сервер бота
	httpServer, err := httpserver.NewBotGateway(ctx, deps.BotServerconfig.HTTPServerConfig, deps.BotHttpHandler)
	if err != nil {
		panic("Failed to create server!")
	}

	// Создаём GRPC-сервер бота (для обработки сообщений по grpc от сервера основной логики)
	grpcBotServer := grpcserver.NewBotGRPCServer(deps.BotGrpcHandler, deps.BotServerconfig.GRPCServerConfig)

	// создаём канал, который бдут реагировать на системные сигналы
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Запуск HTTP сервера бота
	go func() {
		fmt.Printf("🚀 HTTP сервер бота запускается на %s\n", deps.BotServerconfig.HTTPServerConfig.Addr())
		if err := httpServer.Run(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed: %v", err)
		}
	}()

	// запуск GRPC сервера бота
	go func() {
		fmt.Printf("🚀 GRPC сервер бота запускается на %s\n", deps.BotServerconfig.GRPCServerConfig.Addr())
		if err := grpcBotServer.Run(); err != nil {
			log.Fatalf("GRPC Server --- failed: %v", err)
		}
	}()

	// Ожидание сигнала
	<-sigChan
	fmt.Println("\n🛑 Остановка сервера бота...")

	// Graceful shutdown
	shutdownCtx, shutdownCancel := context.WithTimeout(ctx, 30*time.Second)
	defer shutdownCancel()

	// Останавливаем HTTP сервер (ждем текущие запросы)
	fmt.Println("Останавливаем HTTP сервер бота ...")
	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		log.Printf("Error during server shutdown: %v", err)
	}

	// Останавливаем GRPC сервер (ждем текущие запросы)
	fmt.Println("Останавливаем GRPC biz сервер...")
	if err := grpcBotServer.Shutdown(shutdownCtx); err != nil {
		log.Printf("Error during server shutdown: %v", err)
	}

	// Закрываем зависимости при выходе
	err = deps.Close()
	if err != nil {
		log.Printf("Error during resourses closing: %v", err)
	}

	fmt.Println("👋 Bot Серверы остановлены")
}
