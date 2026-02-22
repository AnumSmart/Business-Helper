package grpcclient

import (
	"context"

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
// serverAddr - адрес сервера в формате "host:port" (например "localhost:50051")
func NewBotGrpcClient(serverAddr string) (*BotGrpcClient, error) {
	// grpc.Dial устанавливает соединение с gRPC сервером
	conn, err := grpc.Dial(
		serverAddr,

		// grpc.WithTransportCredentials задает способ аутентификации транспорта
		// insecure.NewCredentials() - отключает TLS/SSL (ТОЛЬКО ДЛЯ РАЗРАБОТКИ!)
		// В продакшене нужно использовать: credentials.NewClientTLSFromFile(...)-----------------------------------------
		grpc.WithTransportCredentials(insecure.NewCredentials()),

		// grpc.WithBlock делает Dial блокирующим - ждем пока соединение установится
		// Без этого опции Dial вернется сразу, даже если сервер недоступен
		grpc.WithBlock(),
	)

	if err != nil {
		return nil, err
	}

	// Создаем клиент для нашего сервиса BotService
	// Это код, сгенерированный protoc из .proto файла
	client := pb.NewBotServiceClient(conn)

	return &BotGrpcClient{
		conn:   conn,
		client: client,
	}, nil
}

// Close закрывает gRPC соединение
// Всегда нужно вызывать при завершении работы приложения
func (c *BotGrpcClient) Close() error {
	return c.conn.Close()
}

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
