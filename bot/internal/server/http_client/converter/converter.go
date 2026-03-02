package converter

import (
	pb "global_models/grpc/bot" // Импорт сгенерированных protobuf структур
)

// ConvertInlineKeyboard конвертирует protobuf inline клавиатуру в формат Telegram API
// Telegram ожидает: {"inline_keyboard": [[{"text": "...", "callback_data": "..."}]]}
func ConvertInlineKeyboard(keyboard *pb.InlineKeyboardMarkup) interface{} {
	// Создаем слайс для рядов кнопок
	result := make([][]map[string]interface{}, len(keyboard.Rows))

	// Проходим по всем рядам
	for i, row := range keyboard.Rows {
		// Создаем слайс для кнопок в текущем ряду
		result[i] = make([]map[string]interface{}, len(row.Buttons))

		// Проходим по всем кнопкам в ряду
		for j, btn := range row.Buttons {
			// Базовая структура кнопки (текст обязателен)
			button := map[string]interface{}{
				"text": btn.Text,
			}

			// Добавляем callback_data если есть (для кнопок с действием)
			if btn.CallbackData != "" {
				button["callback_data"] = btn.CallbackData
			}
			// Добавляем URL если есть (для кнопок-ссылок)
			if btn.Url != "" {
				button["url"] = btn.Url
			}
			result[i][j] = button
		}
	}

	// Оборачиваем в структуру, ожидаемую Telegram API
	return map[string]interface{}{
		"inline_keyboard": result,
	}
}

// ConvertReplyKeyboard конвертирует protobuf обычную клавиатуру в формат Telegram API
// Telegram ожидает: {"keyboard": [[{"text": "..."}]], "resize_keyboard": true}
func ConvertReplyKeyboard(keyboard *pb.ReplyKeyboardMarkup) interface{} {
	// Создаем слайс для рядов кнопок
	result := make([][]map[string]string, len(keyboard.Rows))

	// Проходим по всем рядам
	for i, row := range keyboard.Rows {
		// Создаем слайс для кнопок в ряду
		result[i] = make([]map[string]string, len(row.Buttons))

		// Проходим по всем кнопкам
		for j, btn := range row.Buttons {
			// В обычной клавиатуре только текст (он же будет отправлен как сообщение)
			result[i][j] = map[string]string{
				"text": btn.Text,
			}
		}
	}

	// Возвращаем полную структуру клавиатуры с опциями
	return map[string]interface{}{
		"keyboard":          result,
		"resize_keyboard":   keyboard.ResizeKeyboard,  // Автоподгон размера
		"one_time_keyboard": keyboard.OneTimeKeyboard, // Скрыть после использования
	}
}
