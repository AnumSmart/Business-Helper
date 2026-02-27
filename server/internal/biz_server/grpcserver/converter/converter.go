// конвертеры для работы с сообщениями от grpc клиента
package converter

import (
	pb "global_models/grpc/bot"
	"server/internal/domain"

	"time"
)

// ToIncomingMessage - конвертация из protobuf моделей в доменные модели
func ToIncomingMessage(req *pb.SendMessageRequest) *domain.IncomingMessage {
	// проверка наличи запроса
	if req == nil {
		return nil
	}

	// Создаем "карточку посетителя" на нашем языке
	msg := &domain.IncomingMessage{
		ChatID:     req.ChatId, // Переводим "номер комнаты"
		Text:       req.Text,   // Переводим "текст сообщения"
		ReceivedAt: time.Now(), // Ставим штамп времени
	}

	// Если у посетителя есть с собой клавиатура - переводим и её
	if req.ReplyMarkup != nil {
		msg.ReplyMarkup = toReplyMarkup(req.ReplyMarkup)
	}

	return msg
}

// ToProtoResponse - переводчик ответа с внутреннего языка на внешний
//
// Когда ваши сотрудники (сервисный слой) обработали запрос,
// они пишут ответ на внутреннем языке (domain.MessageResponse).
// Этот переводчик переводит их ответ обратно на язык protobuf,
// чтобы иностранец (клиент) мог его понять.
func ToProtoResponse(resp *domain.MessageResponse) *pb.SendMessageResponse {
	// Если ответа нет - переводим как "пустой ответ с ошибкой"
	if resp == nil {
		return &pb.SendMessageResponse{
			Success: false,
			Error:   "empty response", // "Пустой ответ" на языке клиента
		}
	}

	// Переводим успешность и текст ошибки на язык клиента
	return &pb.SendMessageResponse{
		Success: resp.Success, // Успешно или нет
		Error:   resp.Error,   // Текст ошибки (если была)
	}
}

// ToCallbackLog - переводчик callback-запроса
//
// Callback - это когда пользователь нажал на кнопку.
// К нам приходит информация об этом нажатии на внешнем языке.
// Этот переводчик создает запись в журнале на нашем внутреннем языке:
// - Кто нажал (UserID)
// - Где нажал (ChatID, MessageID)
// - Что нажал (CallbackID, Data)
// - Когда нажал (Timestamp)
//
// Поле ID оставляем пустым - его потом база данных сама присвоит
func ToCallbackLog(pbCallback *pb.CallbackQuery) *domain.CallbackLog {
	if pbCallback == nil {
		return nil // Нет нажатия - ничего не записываем
	}

	// Создаем запись в журнале на нашем языке
	return &domain.CallbackLog{
		CallbackID: pbCallback.Id,        // Уникальный номер нажатия
		UserID:     pbCallback.UserId,    // ID пользователя (кто нажал)
		ChatID:     pbCallback.ChatId,    // ID чата (где нажал)
		MessageID:  pbCallback.MessageId, // ID сообщения (на что нажал)
		Data:       pbCallback.Data,      // Данные кнопки (что именно нажал)
		Timestamp:  time.Now(),           // Время нажатия (когда нажал)
		// ID - оставляем пустым, база сама заполнит
	}
}

// ToDomainMessage преобразует gRPC сообщение от Telegram в доменную модель для БД
//
// Это как переводчик с языка Telegram на язык вашей внутренней системы.
// Telegram присылает сообщение в своем формате (pb.Message), а вы хотите
// сохранить его в базе данных в своем формате (domain.Message).
func ToDomainMessage(pbMsg *pb.Message) *domain.Message {
	// Если сообщения нет - возвращаем nil
	if pbMsg == nil {
		return nil
	}

	// Создаем запись в вашем внутреннем формате
	return &domain.Message{
		MessageID: pbMsg.MessageId,          // ID сообщения в Telegram (как номер чека)
		ChatID:    pbMsg.ChatId,             // ID чата (как номер комнаты)
		UserID:    pbMsg.UserId,             // ID пользователя (кто написал)
		Text:      pbMsg.Text,               // Текст сообщения
		Direction: "incoming",               // Это входящее сообщение (к нам пришло)
		Status:    "received",               // Статус: получено, но еще не обработано
		Timestamp: time.Unix(pbMsg.Date, 0), // Когда написали (переводим из Unix-времени)
		// ID не заполняем - его присвоит база данных при сохранении
	}
}

// ToDomainUser преобразует protobuf пользователя во внутреннюю модель
//
// Telegram прислал нам пользователя в формате protobuf (pb.User),
// а мы хотим работать с ним в сервисном слое в формате domain.User
func ToDomainUser(pbUser *pb.User) *domain.User {
	// Если пользователя нет - возвращаем nil
	if pbUser == nil {
		return nil
	}

	// Переводим с "внешнего" языка на "внутренний"
	return &domain.User{
		ID:        pbUser.Id,
		FirstName: pbUser.FirstName,
		LastName:  pbUser.LastName,
		Username:  pbUser.Username,
		// Дополнительные поля можно добавить позже
	}
}

// ToProtoUser - обратный переводчик (если нужно отправить пользователя обратно)
//
// Может понадобиться, если сервисный слой возвращает пользователя
// для отправки через gRPC
func ToProtoUser(domainUser *domain.User) *pb.User {
	if domainUser == nil {
		return nil
	}

	return &pb.User{
		Id:        domainUser.ID,
		FirstName: domainUser.FirstName,
		LastName:  domainUser.LastName,
		Username:  domainUser.Username,
	}
}
