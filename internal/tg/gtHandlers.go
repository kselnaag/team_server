package tg

import (
	"context"
	"strconv"

	T "team_server/internal/types"

	TG "github.com/go-telegram/bot"
	TGm "github.com/go-telegram/bot/models"
)

func (tg *Tg) startHandler(ctx context.Context, bot *TG.Bot, update *TGm.Update) {
	cfgmsg := T.TS_APP_NAME + "=" + tg.cfg.GetEnvVal(T.TS_APP_NAME) + "\n" +
		T.TS_APP_IP + "=" + tg.cfg.GetEnvVal(T.TS_APP_IP) + "\n" +
		T.TG_BOT_PROXY + "=" + tg.cfg.GetEnvVal(T.TG_BOT_PROXY) + "\n" +
		T.TS_LOG_LEVEL + "=" + tg.cfg.GetEnvVal(T.TS_LOG_LEVEL) + "\n\n"
	var username string
	botUserIDstr := strconv.FormatInt(update.Message.Chat.ID, 10)
	if botUserIDstr == tg.cfg.GetJsonAdmin().TgUserID {
		username = tg.cfg.GetJsonAdmin().Nickname
	} else {
		for _, el := range tg.cfg.GetJsonUsers() {
			if botUserIDstr == el.TgUserID {
				username = el.Nickname
				break
			}
		}
	}
	path := "PATH=" + username + "\n"
	kb := &TGm.ReplyKeyboardMarkup{
		Keyboard:              [][]TGm.KeyboardButton{{{Text: "/ON 🔴"}, {Text: "/start"}, {Text: "/OFF ⚫"}}},
		InputFieldPlaceholder: "Включите или отключите стрим-сервер...",
		ResizeKeyboard:        true,
		IsPersistent:          true,
	}
	_, _ = bot.SendMessage(ctx, &TG.SendMessageParams{
		ChatID:      update.Message.Chat.ID,
		Text:        cfgmsg + path,
		ReplyMarkup: kb,
	})
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

func (tg *Tg) pathHandler(ctx context.Context, bot *TG.Bot, update *TGm.Update) {

}
