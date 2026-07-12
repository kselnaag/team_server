package tg

import (
	"context"
	"strconv"
	"strings"
	T "team_server/internal/types"

	TG "github.com/go-telegram/bot"
	TGm "github.com/go-telegram/bot/models"
)

func (tg *Tg) startHandler(ctx context.Context, bot *TG.Bot, update *TGm.Update) {
	cfgmsg := T.TS_APP_NAME + "=" + tg.cfg.GetEnvVal(T.TS_APP_NAME) + "\n" +
		T.TS_APP_IP + "=" + tg.cfg.GetEnvVal(T.TS_APP_IP) + "\n" +
		T.TG_BOT_PROXY + "=" + tg.cfg.GetEnvVal(T.TG_BOT_PROXY) + "\n" +
		T.TS_LOG_LEVEL + "=" + tg.cfg.GetEnvVal(T.TS_LOG_LEVEL) + "\n"
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
	path := "\nPATH=" + username + "\n"
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

func (tg *Tg) getAuthName(update *TGm.Update) string {
	IDstr := strconv.FormatInt(update.Message.Chat.ID, 10)
	if IDstr == tg.cfg.GetJsonAdmin().TgUserID {
		return tg.cfg.GetJsonAdmin().Nickname
	}
	for _, el := range tg.cfg.GetJsonUsers() {
		if IDstr == el.TgUserID {
			return el.Nickname
		}
	}
	return ""
}

func (tg *Tg) ONHandler(ctx context.Context, bot *TG.Bot, update *TGm.Update) {
	path := tg.getAuthName(update)
	if len(path) != 0 {
		err := tg.msrv.PathInit(path)
		var msg string
		if err != nil {
			msg = "ON " + path + ": " + err.Error()
		} else {
			msg = path + ": ON"
		}
		_, _ = bot.SendMessage(ctx, &TG.SendMessageParams{
			ChatID: update.Message.Chat.ID,
			Text:   msg,
		})
		_, _ = bot.SendMessage(ctx, &TG.SendMessageParams{
			ChatID: tg.cfg.GetJsonChannel(),
			Text:   msg,
		})
	}
}

func (tg *Tg) OFFHandler(ctx context.Context, bot *TG.Bot, update *TGm.Update) {
	path := tg.getAuthName(update)
	if len(path) != 0 {
		err := tg.msrv.PathFini(path)
		var msg string
		if err != nil {
			msg = "OFF " + path + ": " + err.Error()
		} else {
			msg = path + ": OFF"
		}
		_, _ = bot.SendMessage(ctx, &TG.SendMessageParams{
			ChatID: update.Message.Chat.ID,
			Text:   msg,
		})
		_, _ = bot.SendMessage(ctx, &TG.SendMessageParams{
			ChatID: tg.cfg.GetJsonChannel(),
			Text:   msg,
		})
	}
}

func (tg *Tg) getAuthID(name string) string {
	if name == tg.cfg.GetJsonAdmin().Nickname {
		return tg.cfg.GetJsonAdmin().TgUserID
	}
	for _, el := range tg.cfg.GetJsonUsers() {
		if name == el.Nickname {
			return el.TgUserID
		}
	}
	return ""
}

func (tg *Tg) pathHandler(ctx context.Context, bot *TG.Bot, update *TGm.Update) {
	var msg string
	commands := strings.Split(update.Message.Text, " ")
	if len(commands) == 3 {
		path := commands[1]
		pathID := tg.getAuthID(path)
		if (len(path) != 0) && (len(pathID) != 0) && (len(commands[2]) > 1) {
			switch commands[2] {
			case "ON":
				err := tg.msrv.PathInit(path)
				if err != nil {
					msg = "*ON " + path + ": " + err.Error()
				} else {
					msg = "*" + path + ": ON"
				}
				_, _ = bot.SendMessage(ctx, &TG.SendMessageParams{
					ChatID: pathID,
					Text:   msg,
				})
				_, _ = bot.SendMessage(ctx, &TG.SendMessageParams{
					ChatID: tg.cfg.GetJsonChannel(),
					Text:   msg,
				})
				_, _ = bot.SendMessage(ctx, &TG.SendMessageParams{
					ChatID: update.Message.Chat.ID,
					Text:   msg,
				})
				return
			case "OFF":
				err := tg.msrv.PathFini(path)
				if err != nil {
					msg = "*OFF " + path + ": " + err.Error()
				} else {
					msg = "*" + path + ": OFF"
				}
				_, _ = bot.SendMessage(ctx, &TG.SendMessageParams{
					ChatID: pathID,
					Text:   msg,
				})
				_, _ = bot.SendMessage(ctx, &TG.SendMessageParams{
					ChatID: tg.cfg.GetJsonChannel(),
					Text:   msg,
				})
				_, _ = bot.SendMessage(ctx, &TG.SendMessageParams{
					ChatID: update.Message.Chat.ID,
					Text:   msg,
				})
				return
			}
		}
	}
	msg = "ERROR: command string is  \"/path PATH [ON|OFF]\""
	_, _ = bot.SendMessage(ctx, &TG.SendMessageParams{
		ChatID: update.Message.Chat.ID,
		Text:   msg,
	})
}
