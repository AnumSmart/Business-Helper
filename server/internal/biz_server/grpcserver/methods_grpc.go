package grpcserver

import (
	"context"
	"log"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	pb "global_models/grpc/bot"
)

// Входные параметры:
// - ctx: контекст - как таймер разговора (если время вышло - кладем трубку)
// - req: запрос от бота - как звонок от клиента с каким-то вопросом
//
// Возвращает:
// - ответ боту (что нужно сделать: отправить сообщение, показать кнопки и т.д.)
// - ошибку, если что-то пошло не так
func (s *GRPCServer) ProcessUpdate(ctx context.Context, req *pb.UpdateRequest) (*pb.UpdateResponse, error) {
	// Логируем (записываем в журнал), что пришло новое обращение с номером
	// UpdateId - это как уникальный номер обращения, чтобы не потерять
	log.Printf("Processing update %d", req.UpdateId)

	// Определяем тип обновления
	switch {
	// Если в запросе есть сообщение от пользователя (текст, фото, видео)
	case req.Message != nil:
		// Передаем сообщение на обработку в бизнес-логику
		// Как оператор говорит: "Это вопрос про товары, передаю специалисту по товарам"
		return s.Handler.ProcessMessage(ctx, req.Message)

		// Если в запросе есть нажатие на кнопку (callback query)
	case req.CallbackQuery != nil:
		// Передаем нажатие кнопки на обработку в бизнес-логику
		// Как оператор говорит: "Клиент нажал кнопку 'Узнать цену', передаю специалисту по ценам"
		return s.Handler.ProcessCallback(ctx, req.CallbackQuery)
		// Если пришло что-то непонятное (ни сообщение, ни нажатие кнопки)
	default:
		// Возвращаем ошибку - как оператор говорит: "Извините, я не понимаю, что вы хотите"
		// status.Error создает стандартную gRPC ошибку
		// codes.InvalidArgument - код ошибки "неправильный аргумент" (как HTTP 400 Bad Request)
		return nil, status.Error(codes.InvalidArgument, "unsupported update type")
	}
}

// SendMessage - это как отдел исходящих сообщений
// Этот метод вызывается, когда нужно отправить сообщение пользователю
// Например, когда нужно ответить на вопрос или прислать уведомление
func (s *GRPCServer) SendMessage(ctx context.Context, req *pb.SendMessageRequest) (*pb.SendMessageResponse, error) {
	// Логируем, что нужно отправить сообщение в конкретный чат
	// ChatId - это как адрес получателя (уникальный ID чата с пользователем)
	log.Printf("SendMessage request for chat %d", req.ChatId)

	// Передаем запрос в бизнес-логику для отправки сообщения
	// Там решат, какое именно сообщение отправить и с какими кнопками
	//return s.Handler.SendMessage(ctx, req)
	return nil, nil
}
