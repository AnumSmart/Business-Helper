// Пакет telegram предоставляет клиент для взаимодействия с Telegram Bot API
package httpclient

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	pb "global_models/grpc/bot" // Импорт сгенерированных protobuf структур
)

// BotHTTPClient представляет HTTP клиент для Telegram Bot API
// Инкапсулирует все необходимое для взаимодействия с Telegram
type BotHTTPClient struct {
	token   string       // Токен бота (получается от @BotFather)
	Http    *http.Client // HTTP клиент с настроенными таймаутами
	baseURL string       // Базовый URL для API запросов
}

// NewClient создает нового Telegram клиента
// token: токен бота в формате "123456:ABC-DEF1234ghIkl-zyx57W2v1u123ew11"
// Возвращает готовый к работе клиент
func NewClient(token string) *BotHTTPClient {
	return &BotHTTPClient{
		token:   token,
		Http:    &http.Client{Timeout: 10 * time.Second},              // Важно: таймаут защищает от зависания запросов
		baseURL: fmt.Sprintf("https://api.telegram.org/bot%s", token), // Формируем базовый URL согласно документации Telegram
	}
}

// SetWebhook устанавливает webhook URL для получения обновлений от Telegram
// url: публичный HTTPS URL, на который Telegram будет отправлять обновления
// Альтернатива: вместо polling'а (GetUpdates) используем webhook
func (c *BotHTTPClient) SetWebhook(url string) error {
	// Формируем полный URL для метода setWebhook
	// https://api.telegram.org/bot<token>/setWebhook
	webhookURL := fmt.Sprintf("%s/setWebhook", c.baseURL)

	// Создаем тело запроса согласно документации Telegram API
	body := map[string]string{
		"url": url,
	}

	// Сериализуем тело запроса в JSON
	jsonBody, err := json.Marshal(body)
	if err != nil {
		return err
	}

	// Отправляем POST запрос к Telegram API
	resp, err := c.Http.Post(webhookURL, "application/json", bytes.NewBuffer(jsonBody))
	if err != nil {
		return err
	}
	// Важно: закрываем тело ответа, чтобы избежать утечки ресурсов
	defer resp.Body.Close()

	// Структура для парсинга ответа от Telegram
	var result struct {
		Ok          bool   `json:"ok"`          // Успешен ли запрос
		Description string `json:"description"` // Описание ошибки (если не Ok)
	}

	// Декодируем JSON ответ
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return err
	}

	// Проверяем успешность операции
	if !result.Ok {
		return fmt.Errorf("failed to set webhook: %s", result.Description)
	}

	return nil
}

// SendMessage отправляет текстовое сообщение в чат
// chatID: ID получателя (пользователя или группы)
// text: текст сообщения
// replyMarkup: опциональная клавиатура (inline или обычная)
func (c *BotHTTPClient) SendMessage(chatID int64, text string, replyMarkup interface{}) error {
	// Формируем URL для метода sendMessage
	url := fmt.Sprintf("%s/sendMessage", c.baseURL)

	// Создаем тело запроса согласно документации Telegram API
	body := map[string]interface{}{
		"chat_id": chatID, // ID чата (обязательно)
		"text":    text,   // Текст сообщения (обязательно)
	}

	// Добавляем клавиатуру, если она предоставлена
	if replyMarkup != nil {
		body["reply_markup"] = replyMarkup
	}

	// Сериализуем в JSON
	jsonBody, err := json.Marshal(body)
	if err != nil {
		return err
	}

	// Отпрявляем запрос
	resp, err := c.Http.Post(url, "application/json", bytes.NewBuffer(jsonBody))
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	// Простая структура для проверки успешности
	var result struct {
		Ok bool `json:"ok"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return err
	}

	if !result.Ok {
		return fmt.Errorf("failed to send message")
	}

	return nil
}

// SendOutgoingMessages конвертирует gRPC ответы в Telegram формат и отправляет
// messages: массив исходящих сообщений от gRPC сервера
// Это ключевой метод, связывающий gRPC сервер и Telegram API
func (c *BotHTTPClient) SendOutgoingMessages(messages []*pb.OutgoingMessage) error {
	// Проходим по всем сообщениям, которые нужно отправить
	for _, msg := range messages {
		var replyMarkup interface{}

		// Если есть клавиатура, конвертируем из protobuf формата в Telegram формат
		if msg.ReplyMarkup != nil {
			// Определяем тип клавиатуры с помощью type switch
			switch markup := msg.ReplyMarkup.Type.(type) {
			case *pb.ReplyMarkup_InlineKeyboard:
				replyMarkup = convertInlineKeyboard(markup.InlineKeyboard) // Inline клавиатура (кнопки под сообщением)
			case *pb.ReplyMarkup_ReplyKeyboard:
				replyMarkup = convertReplyKeyboard(markup.ReplyKeyboard) // Обычная клавиатура (вместо поля ввода)
			}
		}

		// Отправляем сообщение через Telegram API
		if err := c.SendMessage(msg.ChatId, msg.Text, replyMarkup); err != nil {
			return err
		}
	}
	return nil
}

// convertInlineKeyboard конвертирует protobuf inline клавиатуру в формат Telegram API
// Telegram ожидает: {"inline_keyboard": [[{"text": "...", "callback_data": "..."}]]}
func convertInlineKeyboard(keyboard *pb.InlineKeyboardMarkup) interface{} {
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

// convertReplyKeyboard конвертирует protobuf обычную клавиатуру в формат Telegram API
// Telegram ожидает: {"keyboard": [[{"text": "..."}]], "resize_keyboard": true}
func convertReplyKeyboard(keyboard *pb.ReplyKeyboardMarkup) interface{} {
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
