package tg

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"syscall"
	T "team_server/internal/types"
	"time"

	"github.com/go-telegram/bot"
	TG "github.com/go-telegram/bot"
	TGm "github.com/go-telegram/bot/models"
	"golang.org/x/net/proxy"
)

var _ T.ITG = (*Tg)(nil)

type Tg struct {
	cfg T.ICfg
	log T.ILog
	//msrv T.ImSrv
	ctx context.Context
	bot *TG.Bot
}

func NewTGBot(cfg T.ICfg, log T.ILog /*msrv T.ImSrv*/) *Tg {
	return &Tg{
		cfg: cfg,
		log: log,
		//msrv: msrv,
	}
}

func (tg *Tg) appStop() {
	tg.log.LogDebug("appStop() called")
	proc, _ := os.FindProcess(os.Getpid())
	_ = proc.Signal(syscall.SIGTERM)
}

/* func (tg *Tg) appRestart() {
	tg.log.LogDebug("appRestart() called")
	proc, _ := os.FindProcess(os.Getpid())
	_ = proc.Signal(syscall.SIGHUP)
} */

func (tg *Tg) errorHandler(err error) {
	if errors.Is(err, TG.ErrorBadRequest) {
		err = fmt.Errorf("%s: %w", "ErrorBadRequest 400: ", err)
	}
	if errors.Is(err, TG.ErrorUnauthorized) {
		err = fmt.Errorf("%s: %w", "ErrorUnauthorized 401: ", err)
	}
	if errors.Is(err, TG.ErrorForbidden) {
		err = fmt.Errorf("%s: %w", "ErrorForbidden 403: ", err)
	}
	if errors.Is(err, TG.ErrorNotFound) {
		err = fmt.Errorf("%s: %w", "ErrorNotFound 404: ", err)
	}
	if errors.Is(err, TG.ErrorConflict) {
		err = fmt.Errorf("%s: %w", "ErrorConflict 409: ", err)
	}
	if TG.IsTooManyRequestsError(err) {
		err = fmt.Errorf("TooManyRequestsError 429: retry after %d: %w", err.(*TG.TooManyRequestsError).RetryAfter, err)
	}
	tg.log.LogError(err)
	_, _ = tg.bot.SendMessage(tg.ctx, &TG.SendMessageParams{
		ChatID: tg.cfg.GetJsonAdmin().TgUserID,
		Text:   "TG errorHandler(): " + err.Error(),
	})
}

/* func (tg *Tg) authorized(next TG.HandlerFunc) TG.HandlerFunc {
	return func(ctx context.Context, bot *TG.Bot, update *TGm.Update) {
		if update.Message != nil {
			msg := update.Message
			if (len(msg.Text) > 0) && (msg.Text[0] == '/') {
				usersAutorized := tg.getChatAdmins(tg.cfg.GetJsonAdmin().TgChannelID, update)
				for id := range *usersAutorized {
					if (update.Message.From.ID == id) && (update.Message.Chat.Type == TGm.ChatTypePrivate) {
						next(ctx, bot, update)
						return
					}
				}
				return
			}
			next(ctx, bot, update)
		}
	}
} */

func (tg *Tg) defaultHandler(ctx context.Context, bot *TG.Bot, update *TGm.Update) {}

func (tg *Tg) Start() func(err error) {
	var opts []TG.Option
	if tg.cfg.GetEnvVal(T.TG_BOT_PROXY) == "ON" {
		baseDialer, errP := proxy.SOCKS5("tcp", "127.0.0.1:1080", nil, proxy.Direct)
		if errP != nil {
			tg.log.LogError(fmt.Errorf("TG.Start(): failed to create socks5 dialer: %w", errP))
			return func(err error) {}
		}
		contextDialer, ok := baseDialer.(proxy.ContextDialer)
		if !ok {
			tg.log.LogError(fmt.Errorf("TG.Start(): failed to Type Assert with dialer: dialer does not support context"))
			return func(err error) {}
		}
		customTransport := &http.Transport{DialContext: contextDialer.DialContext}
		customClient := &http.Client{Transport: customTransport}
		opts = []TG.Option{
			TG.WithHTTPClient(10*time.Second, customClient),
			//TG.WithMiddlewares(tg.authorized),
			TG.WithDefaultHandler(tg.defaultHandler),
			TG.WithErrorsHandler(tg.errorHandler),
		}
	} else {
		opts = []TG.Option{
			//TG.WithMiddlewares(tg.authorized),
			TG.WithDefaultHandler(tg.defaultHandler),
			TG.WithErrorsHandler(tg.errorHandler),
		}
	}
	var errBot error
	tg.bot, errBot = TG.New(tg.cfg.GetEnvVal(T.TG_BOT_TOKEN), opts...)
	if nil != errBot {
		tg.log.LogError(fmt.Errorf("TG.Start(): can not create TG bot with error: %w", errBot))
		tg.appStop()
		return func(err error) {}
	}

	var ctxCancelTGbot context.CancelFunc
	tg.ctx, ctxCancelTGbot = context.WithCancel(context.Background())
	_, _ = tg.bot.DeleteMyCommands(tg.ctx, &bot.DeleteMyCommandsParams{})
	_, err := tg.bot.SetMyCommands(tg.ctx, &bot.SetMyCommandsParams{
		Commands: []TGm.BotCommand{{Command: "/start", Description: "Информация + кнопки"}},
		Scope:    &TGm.BotCommandScopeAllPrivateChats{},
	})
	if nil != err {
		tg.log.LogError(fmt.Errorf("TG.Start(): can not create TG bot menu: %w", err))
	}
	_, err = tg.bot.SetChatMenuButton(tg.ctx, &bot.SetChatMenuButtonParams{
		MenuButton: &TGm.MenuButtonCommands{Type: TGm.MenuButtonTypeCommands},
	})
	if err != nil {
		tg.log.LogError(fmt.Errorf("TG.Start(): can not change TG bot menu: %w", err))
	}
	tg.bot.RegisterHandler(TG.HandlerTypeMessageText, "/start", TG.MatchTypeExact, tg.startHandler)
	tg.bot.RegisterHandler(TG.HandlerTypeMessageText, "🔴 ВКЛ", TG.MatchTypeExact, tg.InitHandler)
	tg.bot.RegisterHandler(TG.HandlerTypeMessageText, "⚫ ОТКЛ", TG.MatchTypeExact, tg.FiniHandler)
	tg.bot.RegisterHandler(TG.HandlerTypeMessageText, "/user", TG.MatchTypeCommandStartOnly, tg.userHandler)

	go tg.bot.Start(tg.ctx)
	tg.log.LogInfo("TG.Start(): TG bot started")
	return func(err error) { // TgStop
		ctxCancelTGbot()
		if err != nil {
			tg.log.LogError(fmt.Errorf("TG.Stop(): TG bot stoped with error: %w", err))
		} else {
			tg.log.LogInfo("TG.Stop(): TG bot stoped")
		}
	}
}
