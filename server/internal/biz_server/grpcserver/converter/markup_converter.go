package converter

import (
	pb "global_models/grpc/bot"
	"server/internal/domain"
)

// ----------------------------------------------------------------------------
// toReplyMarkup - переводчик клавиатур с внешнего языка на внутренний
//
// Клавиатуры бывают двух типов:
// - InlineKeyboard - кнопки прямо под сообщением (как ссылки)
// - ReplyKeyboard - обычная клавиатура вместо поля ввода
//
// Эта функция смотрит на тип клавиатуры и переводит её соответствующим образом
func toReplyMarkup(pbMarkup *pb.ReplyMarkup) *domain.ReplyMarkup {
	if pbMarkup == nil {
		return nil // Нет клавиатуры - ничего не переводим
	}

	// Создаем пустую внутреннюю клавиатуру
	markup := &domain.ReplyMarkup{}

	// Смотрим, что за тип клавиатуры нам пришел
	switch v := pbMarkup.Type.(type) {
	case *pb.ReplyMarkup_InlineKeyboard:
		// Это inline клавиатура (кнопки под сообщением)
		// Отправляем её специальному переводчику для inline-клавиатур
		markup.InlineKeyboard = toInlineKeyboard(v.InlineKeyboard)

	case *pb.ReplyMarkup_ReplyKeyboard:
		// Это обычная клавиатура (вместо поля ввода)
		// Отправляем специальному переводчику для обычных клавиатур
		keyboard := toReplyKeyboard(v.ReplyKeyboard)
		markup.Keyboard = keyboard.Keyboard
		markup.ResizeKeyboard = keyboard.ResizeKeyboard
		markup.OneTimeKeyboard = keyboard.OneTimeKeyboard
	}

	return markup // Возвращаем переведенную клавиатуру
}

// toInlineKeyboard - переводчик inline-клавиатуры
//
// Inline-клавиатура - это кнопки, которые прикреплены прямо к сообщению.
// Они могут содержать:
// - callback_data (данные, которые вернутся при нажатии)
// - url (ссылку для открытия)
//
// Клавиатура имеет ряды (rows), а в каждом ряду - кнопки (buttons)
func toInlineKeyboard(pbKeyboard *pb.InlineKeyboardMarkup) [][]domain.InlineButton {
	if pbKeyboard == nil {
		return nil // Нет клавиатуры
	}

	// Создаем такую же структуру, но на нашем языке:
	// result - это срез срезов (ряды кнопок)
	result := make([][]domain.InlineButton, len(pbKeyboard.Rows))

	// Проходим по каждому ряду клавиатуры
	for i, row := range pbKeyboard.Rows {
		// Создаем ряд с нужным количеством кнопок
		result[i] = make([]domain.InlineButton, len(row.Buttons))

		// Проходим по каждой кнопке в ряду
		for j, btn := range row.Buttons {
			// Переводим каждую кнопку:
			// - Текст кнопки
			// - Данные, которые вернутся при нажатии
			// - Ссылку (если есть)
			result[i][j] = domain.InlineButton{
				Text:         btn.Text,
				CallbackData: btn.CallbackData,
				URL:          btn.Url,
			}
		}
	}
	return result // Возвращаем переведенную клавиатуру
}

// toReplyKeyboard - переводчик обычной клавиатуры
//
// Обычная клавиатура - это та, которая появляется вместо поля ввода
// у пользователя в Telegram. У неё есть настройки:
// - resize_keyboard - подгонять размер под кнопки
// - one_time_keyboard - спрятать после использования
func toReplyKeyboard(pbKeyboard *pb.ReplyKeyboardMarkup) *domain.ReplyMarkup {
	if pbKeyboard == nil {
		return nil
	}

	// Создаем клавиатуру с настройками
	markup := &domain.ReplyMarkup{
		ResizeKeyboard:  pbKeyboard.ResizeKeyboard,  // Нужно ли подгонять размер
		OneTimeKeyboard: pbKeyboard.OneTimeKeyboard, // Спрятать после использования
	}

	// Если в клавиатуре есть кнопки
	if len(pbKeyboard.Rows) > 0 {
		// Создаем ряды кнопок
		markup.Keyboard = make([][]domain.Button, len(pbKeyboard.Rows))

		// Проходим по каждому ряду
		for i, row := range pbKeyboard.Rows {
			// Создаем ряд с нужным количеством кнопок
			markup.Keyboard[i] = make([]domain.Button, len(row.Buttons))

			// Проходим по каждой кнопке
			for j, btn := range row.Buttons {
				// У обычной кнопки есть только текст
				markup.Keyboard[i][j] = domain.Button{Text: btn.Text}
			}
		}
	}

	return markup
}

// ----------------------------------------------------------------------------
// ToProtoReplyMarkup преобразует внутреннюю модель клавиатуры в protobuf
//
// Этот конвертер нужен, когда сервисный слой возвращает клавиатуру,
// а мы должны отправить её через gRPC клиенту
func ToProtoReplyMarkup(domainMarkup *domain.ReplyMarkup) *pb.ReplyMarkup {
	if domainMarkup == nil {
		return nil
	}

	// Если есть inline клавиатура
	if len(domainMarkup.InlineKeyboard) > 0 {
		return &pb.ReplyMarkup{
			Type: &pb.ReplyMarkup_InlineKeyboard{
				InlineKeyboard: toProtoInlineKeyboard(domainMarkup.InlineKeyboard),
			},
		}
	}

	// Если есть обычная клавиатура
	if len(domainMarkup.Keyboard) > 0 {
		return &pb.ReplyMarkup{
			Type: &pb.ReplyMarkup_ReplyKeyboard{
				ReplyKeyboard: toProtoReplyKeyboard(domainMarkup),
			},
		}
	}

	return nil
}

// toProtoInlineKeyboard преобразует внутреннюю inline клавиатуру в protobuf
func toProtoInlineKeyboard(domainKeyboard [][]domain.InlineButton) *pb.InlineKeyboardMarkup {
	if len(domainKeyboard) == 0 {
		return nil
	}

	rows := make([]*pb.InlineKeyboardRow, len(domainKeyboard))
	for i, domainRow := range domainKeyboard {
		buttons := make([]*pb.InlineKeyboardButton, len(domainRow))
		for j, domainBtn := range domainRow {
			buttons[j] = &pb.InlineKeyboardButton{
				Text:         domainBtn.Text,
				CallbackData: domainBtn.CallbackData,
				Url:          domainBtn.URL,
			}
		}
		rows[i] = &pb.InlineKeyboardRow{Buttons: buttons}
	}

	return &pb.InlineKeyboardMarkup{Rows: rows}
}

// toProtoReplyKeyboard преобразует внутреннюю обычную клавиатуру в protobuf
func toProtoReplyKeyboard(domainMarkup *domain.ReplyMarkup) *pb.ReplyKeyboardMarkup {
	if len(domainMarkup.Keyboard) == 0 {
		return nil
	}

	rows := make([]*pb.ReplyKeyboardRow, len(domainMarkup.Keyboard))
	for i, domainRow := range domainMarkup.Keyboard {
		buttons := make([]*pb.ReplyKeyboardButton, len(domainRow))
		for j, domainBtn := range domainRow {
			buttons[j] = &pb.ReplyKeyboardButton{Text: domainBtn.Text}
		}
		rows[i] = &pb.ReplyKeyboardRow{Buttons: buttons}
	}

	return &pb.ReplyKeyboardMarkup{
		Rows:            rows,
		ResizeKeyboard:  domainMarkup.ResizeKeyboard,
		OneTimeKeyboard: domainMarkup.OneTimeKeyboard,
	}
}
