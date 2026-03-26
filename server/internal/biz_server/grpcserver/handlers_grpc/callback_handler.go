package handlersgrpc

import (
	"context"
	"fmt"
	pb "global_models/grpc/bot"
	"server/internal/biz_server/grpcserver/converter"
	"server/internal/domain"
)

// создаём структуру контекста колбэка
type callbackContext struct {
	ctx          context.Context
	callback     *domain.CallbackLog
	user         *domain.User
	chatID       int64
	userID       int64
	callbackData string
}

// ProcessCallback - обработка callback от inline клавиатуры
func (b *BizGRPCHandler) ProcessCallback(ctx context.Context, callback *pb.CallbackQuery) (*pb.UpdateResponse, error) {
	// 1. Валидация
	if err := b.validateCallback(callback); err != nil {
		return nil, err
	}

	// 2. Создание контекста callback (сохранение/обновление юзера)
	cbCtx, err := b.buildCallbackContext(ctx, callback)
	if err != nil {
		return nil, err
	}

	// 3. Сохранение callback (сначала сохраняем)
	b.saveCallback(cbCtx)

	// 4. Логирование (теперь можем использовать сохраненные данные)
	b.logCallback(cbCtx)

	// 5. Обработка callback и формирование ответа
	return b.handleCallbackAction(cbCtx)

}

// метод валидации, проверячем колбэк на nil
func (b *BizGRPCHandler) validateCallback(callback *pb.CallbackQuery) error {
	if callback == nil {
		return fmt.Errorf("callback is nil")
	}
	return nil
}

// метод создания контектса колбэка
func (b *BizGRPCHandler) buildCallbackContext(ctx context.Context, callback *pb.CallbackQuery) (*callbackContext, error) {
	// конвертируем grpc колбэк в доменную структуру
	callbackLog := converter.ToCallbackLog(callback)

	// сохраняем/обновляем пользователя
	user, err := b.Service.Users.RegisterOrUpdate(ctx,
		callbackLog.MessageID,
		callbackLog.UserFirstName,
		callbackLog.UserLastName,
		callbackLog.UserNickName)
	if err != nil {
		fmt.Printf("⚠️ Failed to save user: %v", err)
	}

	// заполняем структуру контекста колбэка
	return &callbackContext{
		ctx:          ctx,
		callback:     callbackLog,
		user:         user,
		chatID:       callback.ChatId,
		userID:       callback.UserId,
		callbackData: callback.Data,
	}, nil
}

// сохраняем колбэк в БД
func (b *BizGRPCHandler) saveCallback(cbCtx *callbackContext) {

	// Сохраняем callback в БД
	if err := b.Service.Messages.CheckAndSaveCallBack(cbCtx.ctx, cbCtx.callback); err != nil {
		fmt.Printf("⚠️ Failed to save callback: %v", err)
	}
}

// метод для логирования колбэка
func (b *BizGRPCHandler) logCallback(cbCtx *callbackContext) {
	// Используем сохраненные данные для логирования
	userName := "unknown"
	if cbCtx.user != nil {
		userName = cbCtx.user.Username
		if userName == "" {
			userName = fmt.Sprintf("%s %s", cbCtx.user.FirstName, cbCtx.user.LastName)
		}
	}

	fmt.Printf("🔘 Callback processed: ID=%s, User=%s (ID=%d), ChatID=%d, Data=%s",
		cbCtx.callback.ID,
		userName,
		cbCtx.userID,
		cbCtx.chatID,
		cbCtx.callbackData)
}

// метод возвращения ответа в grpc формате
func (b *BizGRPCHandler) handleCallbackAction(cbCtx *callbackContext) (*pb.UpdateResponse, error) {
	// формируем успешный вариант (оптимистичный прогноз)
	response := &pb.UpdateResponse{Success: true}

	// объявляем кастомный тип для работы с хэндлерами
	type handler func(*callbackContext) *pb.UpdateResponse

	// Инициализируем (через литерал) map для обработки команд (ключ - строка, значение - функция)
	handlers := map[string]handler{
		"help":          b.handleHelpCallback,
		"lookup":        b.handleLookupCallback,
		"menu":          b.handleMenuCallback,
		"contacted_yes": b.handleContactedYes,
		"contacted_no":  b.handleContactedNo,
	}

	// если в мапе есть такой обработчик - то вызываем его и возвращаем результат
	if handler, exists := handlers[cbCtx.callbackData]; exists {
		result := handler(cbCtx)
		fmt.Printf("✅ Callback '%s' handled successfully for user %d",
			cbCtx.callbackData, cbCtx.userID)
		return result, nil
	}

	// Обработка неизвестной команды
	// заполняем response
	response.Messages = append(response.Messages, &pb.OutgoingMessage{
		ChatId: cbCtx.chatID,
		Text:   fmt.Sprintf("❓ Неизвестная команда: %s", cbCtx.callbackData),
	})

	fmt.Printf("⚠️ Unknown callback command: %s from user %d",
		cbCtx.callbackData, cbCtx.userID)

	return response, nil
}

// обработчик для колбэка "help"
func (b *BizGRPCHandler) handleHelpCallback(cbCtx *callbackContext) *pb.UpdateResponse {
	return &pb.UpdateResponse{
		Success: true,
		Messages: []*pb.OutgoingMessage{
			{
				ChatId: cbCtx.chatID,
				Text:   "🤖 Я бот-помощник. Доступные команды:\n/help - помощь\n/menu - главное меню",
			},
		},
	}
}

// обработчик для колбэка "lookup"
func (b *BizGRPCHandler) handleLookupCallback(cbCtx *callbackContext) *pb.UpdateResponse {
	btns := [][]domain.InlineButton{
		{
			{Text: "🔗 Перейти в Instagram", URL: "https://www.instagram.com/..."},
		},
		{
			{Text: "✅ Уже связался", CallbackData: "contacted_yes"},
			{Text: "❌ Пока не готов", CallbackData: "contacted_no"},
		},
	}

	// создаём клавиатуру на базе доменной модели
	replyMarkup := &domain.ReplyMarkup{InlineKeyboard: btns}

	return &pb.UpdateResponse{
		Success: true,
		Messages: []*pb.OutgoingMessage{
			{
				ChatId:      cbCtx.chatID,
				Text:        "📸 Вот ссылка на Instagram аккаунт мастера:\nПосле просмотра, пожалуйста, выберите вариант:",
				ReplyMarkup: converter.ToProtoReplyMarkup(replyMarkup),
			},
		},
	}
}

// обработчик для колбэка "menu"
func (b *BizGRPCHandler) handleMenuCallback(cbCtx *callbackContext) *pb.UpdateResponse {
	btns := [][]domain.InlineButton{
		{
			{Text: "🆘 Помощь", CallbackData: "help"},
			{Text: "🔍 Ознакомиться", CallbackData: "lookup"},
		},
	}

	replyMarkup := &domain.ReplyMarkup{InlineKeyboard: btns}

	return &pb.UpdateResponse{
		Success: true,
		Messages: []*pb.OutgoingMessage{
			{
				ChatId:      cbCtx.chatID,
				Text:        "🏠 Главное меню",
				ReplyMarkup: converter.ToProtoReplyMarkup(replyMarkup),
			},
		},
	}
}

// обработчик для колбэка "contacted_yes"
func (b *BizGRPCHandler) handleContactedYes(cbCtx *callbackContext) *pb.UpdateResponse {
	// Здесь можно добавить дополнительную логику, например:
	// - Сохранить в БД факт согласия на связь
	// - Отправить уведомление администратору
	// - Запустить бизнес-процесс

	return &pb.UpdateResponse{
		Success: true,
		Messages: []*pb.OutgoingMessage{
			{
				ChatId: cbCtx.chatID,
				Text:   "✅ Отлично! Я передам ваши контакты мастеру. Ожидайте связи в ближайшее время.",
			},
		},
	}
}

// обработчик для колбэка "contacted_no"
func (b *BizGRPCHandler) handleContactedNo(cbCtx *callbackContext) *pb.UpdateResponse {
	return &pb.UpdateResponse{
		Success: true,
		Messages: []*pb.OutgoingMessage{
			{
				ChatId: cbCtx.chatID,
				Text:   "💭 Жаль! Если передумаете, просто нажмите /start, чтобы вернуться в меню.",
			},
		},
	}
}
