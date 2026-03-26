package servicegrpc

import (
	"fmt"
	"server/internal/domain"
)

// ========== Response Generator Service ==========
type ResponseGenerator interface {
	GenerateReply(text string, user *domain.User) string
	CreateTestKeyboard() *domain.ReplyMarkup
	CreateWelcomeReplyKeyboard() *domain.ReplyMarkup
	CreateTextRespKeyBoard(text string) *domain.ReplyMarkup
}

// структура сервиса генератора ответов
type responseGenerator struct{}

// конструктор сервиса генератора ответов
func NewResponseGenerator() ResponseGenerator {
	return &responseGenerator{}
}

// generateReply генерирует ответ на сообщение
func (g *responseGenerator) GenerateReply(text string, user *domain.User) string {
	if text == "🏠 Главное меню" {
		return "Вы вернулись в главное меню. Пожалуйста, выберите действие:"
	}

	if text == "❓ Помощь" {
		return "Этот бот создан, чтобы облегчить вам жизнь"
	}

	return fmt.Sprintf("Пришло непредвиденное сообщение: %s", text)
}

// CreateTestReplyKeyboard создает тестовую обычную клавиатуру
// (альтернативный пример для полноты)
func (s *responseGenerator) CreateTestKeyboard() *domain.ReplyMarkup {
	return &domain.ReplyMarkup{
		Keyboard: [][]domain.Button{
			// Первый ряд
			{
				{Text: "Кнопка 1"},
				{Text: "Кнопка 2"},
			},
			// Второй ряд
			{
				{Text: "Отмена"},
			},
		},
		ResizeKeyboard:  true, // Подогнать размер под кнопки
		OneTimeKeyboard: true, // Спрятать после использования
	}
}

// CreateWelcomeReplyKeyboard создает приветственную клавиатуру
func (s *responseGenerator) CreateWelcomeReplyKeyboard() *domain.ReplyMarkup {
	return &domain.ReplyMarkup{
		InlineKeyboard: [][]domain.InlineButton{
			// Первый ряд
			{
				{
					Text: "Поисковик",
					URL:  "https://www.google.com/",
				},
			},
		},
		ResizeKeyboard:  true, // Подогнать размер под кнопки
		OneTimeKeyboard: true, // Спрятать после использования
	}
}

// метод для создания клавиатуры в зависимости от текста сообщения
func (s *responseGenerator) CreateTextRespKeyBoard(text string) *domain.ReplyMarkup {
	switch text {
	case "🏠 Главное меню":
		return s.getMainMenuKeyboard()
	}

	return s.getMainMenuKeyboard()
}

// метод для добавления inline клавиатуры для главного меню
func (s *responseGenerator) getMainMenuKeyboard() *domain.ReplyMarkup {
	return &domain.ReplyMarkup{
		InlineKeyboard: [][]domain.InlineButton{
			{
				{Text: "📚 Ознакомиться", CallbackData: "lookup"},
			},
		},
	}
}
