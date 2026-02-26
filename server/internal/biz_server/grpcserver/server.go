package grpcserver

import (
	"context"
	"fmt"
	"log"
	"net"
	"pkg/configs"

	"server/internal/interfaces"

	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
	"google.golang.org/grpc/reflection"

	pb "global_models/grpc/bot" // Импортируем сгенерированные protobuf - это как контракт, по которому бот и сервер будут общаться
	"global_models/interf"
)

// GRPCServer - gRPC сервер для приема запросов от бота
type GRPCServer struct {
	pb.UnimplementedBotServiceServer                                 // Встраиваем для обратной совместимости
	server                           *grpc.Server                    // Сам сервер, который слушает входящие подключения
	Handler                          interfaces.GRPCHandlerInterface // Бизнес-логика для сообщений (интерфейс из сервисного слоя)
	config                           *configs.GRPCServerConfig       // конфиг grpc сервера
}

// NewGRPCServer создает новый gRPC сервер (конструктор), возвращает глобальный интерфейс
func NewGRPCServer(handler interfaces.GRPCHandlerInterface, conf *configs.GRPCServerConfig) interf.GRPCInterface {
	return &GRPCServer{
		Handler: handler,
		config:  conf,
	}
}

// Run - запускает сервер, чтобы он начал принимать заказы от бота
func (s *GRPCServer) Run() error {
	// Говорим операционной системе: "Слушай входящие подключения на таком-то порту"
	lis, err := net.Listen("tcp", ":"+s.config.Port)
	if err != nil {
		// Если не получилось открыть порт (например, он уже занят) - сообщаем об ошибке
		return fmt.Errorf("failed to listen: %w", err)
	}

	// Настройки keepalive для надежности соединения
	keepaliveParams := keepalive.ServerParameters{
		MaxConnectionIdle:     s.config.MaxConnectionIdle,     // Если клиент молчит 15 минут - можно закрыть соединение
		MaxConnectionAge:      s.config.MaxConnectionAge,      // Максимальное время жизни соединения - 30 минут
		MaxConnectionAgeGrace: s.config.MaxConnectionAgeGrace, // Даем 5 минут на завершение текущих дел перед закрытием
		Time:                  s.config.KeepaliveTime,         // Каждые 5 минут проверяем, жив ли клиент
		Timeout:               s.config.KeepaliveTimeout,      // Ждем ответ 20 секунд, если не отвечает - считаем отключившимся
	}

	// Создаем gRPC сервер с опциями
	s.server = grpc.NewServer(
		grpc.KeepaliveParams(keepaliveParams),        // Добавляем проверки соединения
		grpc.MaxRecvMsgSize(s.config.MaxRecvMsgSize), // Максимальный размер принимаемого сообщения - 10 МБ
		grpc.MaxSendMsgSize(s.config.MaxSendMsgSize), // Максимальный размер отправляемого сообщения - тоже 10 МБ
	)

	// Регистрируем наш сервис - говорим: "Этот сервер умеет работать с ботом по таким-то правилам"
	pb.RegisterBotServiceServer(s.server, s)

	// Регистрируем reflection для инструментов отладки (grpcurl и т.д.)
	reflection.Register(s.server)

	log.Printf("gRPC server listening on :%s", s.config.Port)

	// Запускаем сервер в бесконечный цикл приема сообщений
	// Serve - блокирующая операция, выполняется пока сервер не остановят
	return s.server.Serve(lis)
}

// Shutdown - аккуратно останавливает сервер, давая завершить текущие задания
func (s *GRPCServer) Shutdown(ctx context.Context) error {
	stopped := make(chan struct{}) // Создаем канал, который сообщит, когда сервер остановится

	// Запускаем горутину для graceful shutdown
	go func() {
		// GracefulStop - вежливо просим сервер остановиться
		// Новые подключения не принимаем, но текущие дообслуживаем
		s.server.GracefulStop()
		close(stopped) // Сообщаем, что остановка завершена
	}()

	// Ждем или завершения остановки, или истечения времени в контектсе
	select {
	case <-ctx.Done():
		s.server.Stop() // Грубо останавливаем все соединения
		return ctx.Err()
	case <-stopped:
		log.Println("gRPC server shutdown completed")
		return nil
	}
}
