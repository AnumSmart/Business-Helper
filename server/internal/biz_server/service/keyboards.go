package service

import "server/internal/domain"

// метод для добавления inline клавиатуры для главного меню
func (s *BizService) getMainMenuKeyboard() *domain.ReplyMarkup {
	return &domain.ReplyMarkup{
		InlineKeyboard: [][]domain.InlineButton{
			{
				{Text: "📚 Ознакомиться", CallbackData: "lookup"},
			},
		},
	}
}
