package handlers

import (
	"strings"

	tele "gopkg.in/telebot.v4"
)

// хэндлер для обработки команды /start от телеграмм бота в polling режиме
func (h *BotHttpHandler) HandleBotStart(c tele.Context) error {

	args := strings.Fields(c.Text())

	var welcomeMsg string

	if len(args) > 1 {
		switch args[1] {
		case "menu":
			welcomeMsg = "🏠 Главное меню\n\nДобро пожаловать! Я бот-ассистент. Выберите действие:"
		case "help":
			welcomeMsg = "🆘 Справка\n\nЯ могу помочь вам:\n• Ознакомиться с примерами работ\n• Ответить на вопросы\n• Связаться с мастером"
		default:
			welcomeMsg = "Добро пожаловать! Я бот-ассистент. Вы можете ознакомиться с примерами работ по дизайну и связаться с мастером."
		}
	} else {
		welcomeMsg = "Добро пожаловать! Я бот-ассистент. Вы можете ознакомиться с примерами работ по дизайну и связаться с мастером."
	}

	// Отправляем одно сообщение с reply клавиатурой, при нажатии на одну из этих кнопок, посылается текстовое сообщение
	replyMarkup := &tele.ReplyMarkup{
		ResizeKeyboard: true,
		InlineKeyboard: [][]tele.InlineButton{
			{
				{
					Text: "📚 Ознакомиться",
					Data: "lookup",
				},
			},
		},
		ReplyKeyboard: [][]tele.ReplyButton{
			{
				{Text: "🏠 Главное меню"},
				{Text: "❓ Помощь"},
			},
		},
	}

	return c.Send(welcomeMsg, replyMarkup)
}
