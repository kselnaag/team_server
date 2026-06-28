package tg

import (
	"context"
	"log"

	TG "github.com/go-telegram/bot"
	TGm "github.com/go-telegram/bot/models"
)

func (tg *Tg) startHandler(ctx context.Context, bot *TG.Bot, update *TGm.Update) {
	kb := &TGm.ReplyKeyboardMarkup{
		Keyboard:              [][]TGm.KeyboardButton{{{Text: "🔴 ВКЛ"}, {Text: "⚫ ОТКЛ"}}},
		InputFieldPlaceholder: "Включите или отключите стрим-сервер...",
		ResizeKeyboard:        true,
		IsPersistent:          false,
	}
	_, err := bot.SendMessage(ctx, &TG.SendMessageParams{
		ChatID:      update.Message.Chat.ID,
		Text:        "Привет! Добро пожаловать. Выберите один из вариантов ниже:",
		ReplyMarkup: kb,
	})
	if err != nil {
		log.Printf("Ошибка отправки клавиатуры: %v", err)
	}
}

func (tg *Tg) InitHandler(ctx context.Context, bot *TG.Bot, update *TGm.Update) {
	_, _ = bot.SendMessage(ctx, &TG.SendMessageParams{
		ChatID: update.Message.Chat.ID,
		Text:   "INIT",
	})
}

func (tg *Tg) FiniHandler(ctx context.Context, bot *TG.Bot, update *TGm.Update) {
	_, _ = bot.SendMessage(ctx, &TG.SendMessageParams{
		ChatID: update.Message.Chat.ID,
		Text:   "FINI",
	})
}

func (tg *Tg) userHandler(ctx context.Context, bot *TG.Bot, update *TGm.Update) {

}
