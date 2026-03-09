package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"server/internal/biz_server/grpcserver"
	"server/internal/biz_server/httpserver"
	"server/internal/dependencies"
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
	httpServer, err := httpserver.NewBizServer(ctx, deps.BizConfig.HTTPServerConf, deps.BizHTTPHandler)
	if err != nil {
		panic("Failed to create server!")
	}

	// Создаём GRPC-сервер
	grpcServer := grpcserver.NewGRPCServer(deps.BizGRPCHandler, deps.BizConfig.GRPCServerConf)

	// создаём канал, который бдут реагировать на системные сигналы
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Запуск HTTP сервера
	go func() {
		fmt.Printf("🚀 HTTP сервер основной логики запускается на %s\n", deps.BizConfig.HTTPServerConf.Addr())
		if err := httpServer.Run(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("HTTP Server --- failed: %v", err)
		}
	}()

	// запуск GRPC сервера
	go func() {
		fmt.Printf("🚀 GRPC сервер основной логики запускается на %s\n", deps.BizConfig.GRPCServerConf.Addr())
		if err := grpcServer.Run(); err != nil {
			log.Fatalf("GRPC Server --- failed: %v", err)
		}
	}()

	// Ожидание сигнала
	<-sigChan
	fmt.Println("\n🛑 Остановка сервера biz...")

	// Graceful shutdown - контекст с отменой, чтобы дать 30 сек на завершение 2х серверов
	shutdownCtx, shutdownCancel := context.WithTimeout(ctx, 30*time.Second)
	defer shutdownCancel()

	// Останавливаем HTTP сервер (ждем текущие запросы)
	fmt.Println("Останавливаем HTTP biz сервер...")
	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		log.Printf("Error during server shutdown: %v", err)
	}

	// Останавливаем GRPC сервер (ждем текущие запросы)
	fmt.Println("Останавливаем GRPC biz сервер...")
	if err := grpcServer.Shutdown(shutdownCtx); err != nil {
		log.Printf("Error during server shutdown: %v", err)
	}

	// Закрываем зависимости при выходе
	err = deps.Close()
	if err != nil {
		log.Printf("Error during resourses closing: %v", err)
	}

	fmt.Println("👋 Серверы остановлены")

}
