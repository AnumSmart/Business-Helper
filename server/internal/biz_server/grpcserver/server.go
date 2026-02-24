package grpcserver

import (
	"context"
	"fmt"
	"log"
	"net"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
	"google.golang.org/grpc/reflection"

	pb "global_models/grpc/bot" // Импортируем сгенерированные protobuf - это как контракт, по которому бот и сервер будут общаться
	"server/internal/biz_server/service"
)

// GRPCServer - gRPC сервер для приема запросов от бота
type GRPCServer struct {
	pb.UnimplementedBotServiceServer                                 // Встраиваем для обратной совместимости
	server                           *grpc.Server                    // Сам сервер, который слушает входящие подключения
	messageService                   service.MessageServiceInterface // Бизнес-логика для сообщений (интерфейс из сервисного слоя)
	port                             string                          // Порт, на котором сервер будет работать (например, 50051)
}

// NewGRPCServer создает новый gRPC сервер (конструктор)
func NewGRPCServer(messageService service.MessageServiceInterface, port string) *GRPCServer {
	return &GRPCServer{
		messageService: messageService,
		port:           port,
	}
}

// Run - запускает сервер, чтобы он начал принимать заказы от бота
func (s *GRPCServer) Run() error {
	// Говорим операционной системе: "Слушай входящие подключения на таком-то порту"
	lis, err := net.Listen("tcp", ":"+s.port)
	if err != nil {
		// Если не получилось открыть порт (например, он уже занят) - сообщаем об ошибке
		return fmt.Errorf("failed to listen: %w", err)
	}

	// Настройки keepalive для надежности соединения
	keepaliveParams := keepalive.ServerParameters{
		MaxConnectionIdle:     15 * time.Minute, // Если клиент молчит 15 минут - можно закрыть соединение
		MaxConnectionAge:      30 * time.Minute, // Максимальное время жизни соединения - 30 минут
		MaxConnectionAgeGrace: 5 * time.Minute,  // Даем 5 минут на завершение текущих дел перед закрытием
		Time:                  5 * time.Minute,  // Каждые 5 минут проверяем, жив ли клиент
		Timeout:               20 * time.Second, // Ждем ответ 20 секунд, если не отвечает - считаем отключившимся
	}

	// Создаем gRPC сервер с опциями
	s.server = grpc.NewServer(
		grpc.KeepaliveParams(keepaliveParams), // Добавляем проверки соединения
		grpc.MaxRecvMsgSize(1024*1024*10),     // Максимальный размер принимаемого сообщения - 10 МБ
		grpc.MaxSendMsgSize(1024*1024*10),     // Максимальный размер отправляемого сообщения - тоже 10 МБ
	)

	// Регистрируем наш сервис - говорим: "Этот сервер умеет работать с ботом по таким-то правилам"
	pb.RegisterBotServiceServer(s.server, s)

	// Регистрируем reflection для инструментов отладки (grpcurl и т.д.)
	reflection.Register(s.server)

	log.Printf("gRPC server listening on :%s", s.port)

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
