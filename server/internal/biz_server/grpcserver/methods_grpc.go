package grpcserver

import (
	"context"
	"fmt"
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
	// Логируем факт получения обновления
	// UpdateId - это как номер обращения в техподдержку
	log.Printf("Processing update %d", req.UpdateId)

	// Создаем временные хранилища для:
	// - responses: сюда складываем успешные ответы от обработчиков
	// - errors: сюда складываем ошибки, если что-то пошло не так
	var responses []*pb.UpdateResponse
	var errors []error

	// ШАГ 1: Обрабатываем сообщение от пользователя (если оно есть)
	// Например, пользователь написал "Привет!" или прислал фото
	if req.Message != nil {
		// Вызываем специалиста по сообщениям (ProcessMessage)
		// Передаем ему само сообщение для анализа
		resp, err := s.Handler.ProcessMessage(ctx, req.Message)

		if err != nil {
			// Если специалист вернул ошибку - записываем её
			// Но продолжаем работу (может быть, еще есть callback)
			errors = append(errors, err)
		} else if resp != nil {
			// Если специалист успешно обработал - сохраняем ответ
			responses = append(responses, resp)
		}
	}

	// ШАГ 2: Обрабатываем нажатие на кнопку (если оно есть)
	// Например, пользователь нажал кнопку "Узнать цену"
	if req.CallbackQuery != nil {
		// Вызываем специалиста по callback'ам
		resp, err := s.Handler.ProcessCallback(ctx, req.CallbackQuery)

		if err != nil {
			errors = append(errors, err)
		} else if resp != nil {
			responses = append(responses, resp)
		}
	}

	// ШАГ 3: Проверяем, что нам вообще что-то прислали
	// Если нет ни сообщения, ни callback, и ни одной ошибки -
	// значит пришел пустой запрос (такого быть не должно)
	if len(responses) == 0 && len(errors) == 0 {
		// Возвращаем gRPC ошибку с кодом "неверный аргумент"
		// Клиент поймет, что запрос был некорректным
		return nil, status.Error(codes.InvalidArgument, "no message or callback provided")
	}

	// ШАГ 4: Собираем финальный ответ
	// Начинаем с оптимистичного ответа "все хорошо"
	finalResp := &pb.UpdateResponse{Success: true}

	// Проходим по всем успешным ответам
	for _, resp := range responses {
		// Если хоть один обработчик вернул false - весь ответ false
		if !resp.Success {
			finalResp.Success = false
		}
		// Собираем все исходящие сообщения от всех обработчиков
		// Например, на сообщение может быть 2 ответа: "Спасибо!" и рекламный баннер
		finalResp.Messages = append(finalResp.Messages, resp.Messages...)
	}

	// ШАГ 5: Если были ошибки - добавляем их в ответ
	if len(errors) > 0 {
		finalResp.Success = false
		// Объединяем все ошибки в одну строку
		// Клиент увидит что-то вроде: "errors: [ошибка1 ошибка2]"
		finalResp.Error = fmt.Sprintf("errors: %v", errors)
	}

	// ШАГ 6: Возвращаем собранный ответ
	// Клиент получит:
	// - Флаг успеха (true/false)
	// - Список сообщений для отправки пользователю
	// - Текст ошибки (если были проблемы)
	return finalResp, nil
}

// SendMessage - это реализация метода на стороне grpc сервера
// если со стороны grpc  клиента был вызван этот метод, будет вызываться эта реализация
func (s *GRPCServer) SendMessage(ctx context.Context, req *pb.SendMessageRequest) (*pb.SendMessageResponse, error) {
	// Логируем, что нужно отправить сообщение в конкретный чат
	// ChatId - это как адрес получателя (уникальный ID чата с пользователем)
	log.Printf("SendMessage request for chat %d", req.ChatId)

	// Передаем запрос в бизнес-логику для обработки принятого сообщения от grpc клиента на стороне бота
	// проведём проверки и передадим в сервисный слой, чтобы там решить куда дальше посылать ответ (если нужно будет)
	return s.Handler.ProcessIncomingMsg(ctx, req)
}
