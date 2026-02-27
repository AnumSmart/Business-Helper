package grpcclient

import (
	"fmt"
	"pkg/configs"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	// Импортируем сгенерированный из proto файла пакет
	// pb - это псевдоним (alias) для удобства использования
	pb "global_models/grpc/bot"
)

// BotGrpcClient представляет gRPC клиент для сервиса бота
// Инкапсулирует соединение и сгенерированный клиент
type BotGrpcClient struct {
	conn   *grpc.ClientConn    // Физическое соединение с сервером
	client pb.BotServiceClient // Сгенерированный клиент для вызова методов
}

// NewBotGrpcClient создает новый gRPC клиент и устанавливает соединение с сервером
// serverAddr - адрес сервера в формате "host:port" (например "localhost:50052")
func NewBotGrpcClient(config *configs.GRPCClientConfig) (*BotGrpcClient, error) {
	// Создаем клиент
	conn, err := grpc.NewClient(
		config.Addr(),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create gRPC client: %w", err)
	}

	// Инициируем соединение
	conn.Connect()

	// Проверяем состояние сразу
	client := pb.NewBotServiceClient(conn)

	return &BotGrpcClient{
		conn:   conn,
		client: client,
	}, nil
}

// Close закрывает gRPC соединение
// Всегда нужно вызывать при завершении работы приложения
func (c *BotGrpcClient) ShutDown() error {
	return c.conn.Close()
}

// Необходимо будет использовать только нужные методы grpc сервера

/*
// ProcessUpdate отправляет запрос на обработку обновления от Telegram
func (c *BotGrpcClient) ProcessUpdate(ctx context.Context, req *pb.UpdateRequest) (*pb.UpdateResponse, error) {
	// Вызываем сгенерированный метод клиента
	// Запрос автоматически сериализуется в protobuf и отправляется по gRPC
	return c.client.ProcessUpdate(ctx, req)
}

// SendMessage отправляет запрос на отправку сообщения от бота
func (c *BotGrpcClient) SendMessage(ctx context.Context, req *pb.SendMessageRequest) (*pb.SendMessageResponse, error) {
	return c.client.SendMessage(ctx, req)
}
*/
