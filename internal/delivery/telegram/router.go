package telegram

import (
	"context"
	"fmt"
	"strings"
	"sync"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/m4xvel/monetych_bot/internal/logger"
)

type HandlerFunc func(ctx context.Context, msg *tgbotapi.Message)
type CallbackHandlerFunc func(ctx context.Context, cb *tgbotapi.CallbackQuery)

type Router struct {
	commandHandlers  map[string]HandlerFunc
	callbackHandlers map[string]CallbackHandlerFunc
	messageHandler   HandlerFunc

	mu    sync.Mutex
	locks map[string]struct{}
}

func NewRouter() *Router {
	return &Router{
		commandHandlers:  make(map[string]HandlerFunc),
		callbackHandlers: make(map[string]CallbackHandlerFunc),
		locks:            make(map[string]struct{}),
	}
}

func (r *Router) RegisterCommand(cmd string, handler HandlerFunc) {
	r.commandHandlers[cmd] = handler
}

func (r *Router) RegisterCallback(prefix string, handler CallbackHandlerFunc) {
	r.callbackHandlers[prefix] = handler
}

func (r *Router) RegisterMessageHandler(handler HandlerFunc) {
	r.messageHandler = handler
}

func (r *Router) Route(ctx context.Context, upd tgbotapi.Update) {
	switch {

	case upd.Message != nil:
		msg := upd.Message

		if msg.IsCommand() {
			cmd := msg.Command()

			logger.Log.Infow("command received",
				"user_id", msg.From.ID,
				"username", msg.From.UserName,
				"command", cmd,
			)

			if handler, ok := r.commandHandlers[cmd]; ok {
				handler(ctx, msg)
				return
			}

			logger.Log.Warnw("no command handler",
				"user_id", msg.From.ID,
				"command", cmd,
			)

			return
		}

		if r.messageHandler != nil {
			logger.Log.Debugw("message received",
				"user_id", msg.From.ID,
				"text", msg.Text,
			)

			r.messageHandler(ctx, msg)
			return
		}

	case upd.CallbackQuery != nil:
		cb := upd.CallbackQuery
		lockKey := buildLockKey(cb)

		logger.Log.Infow("callback received",
			"user_id", cb.From.ID,
			"username", cb.From.UserName,
		)

		r.mu.Lock()
		if _, exists := r.locks[lockKey]; exists {
			r.mu.Unlock()

			logger.Log.Warnw("callback ignored (already handled)",
				"user_id", cb.From.ID,
				"data", cb.Data,
			)

			return
		}
		r.locks[lockKey] = struct{}{}
		r.mu.Unlock()

		for prefix, h := range r.callbackHandlers {
			if strings.HasPrefix(cb.Data, prefix) {
				h(ctx, cb)
				return
			}
		}

		logger.Log.Warnw("no callback handler",
			"user_id", cb.From.ID,
			"data", cb.Data,
		)

	default:
		logger.Log.Debugw("unknown update received")
	}
}

func buildLockKey(cb *tgbotapi.CallbackQuery) string {
	return fmt.Sprintf(
		"%d:%d:%d:%s",
		cb.From.ID,
		cb.Message.Chat.ID,
		cb.Message.MessageID,
		cb.Data,
	)
}
