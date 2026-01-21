package telegram

import (
	"context"
	"fmt"
	"sync"
	"time"

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
	locks map[string]time.Time
}

func NewRouter() *Router {
	r := &Router{
		commandHandlers:  make(map[string]HandlerFunc),
		callbackHandlers: make(map[string]CallbackHandlerFunc),
		locks:            make(map[string]time.Time),
	}

	go r.cleanup()
	return r
}

const lockTTL = 15 * time.Second

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
		action := extractPrefix(cb.Data)
		lockKey := buildLockKey(cb, action)

		logger.Log.Infow("callback received",
			"user_id", cb.From.ID,
			"username", cb.From.UserName,
			"action", action,
		)

		r.mu.Lock()
		if t, ok := r.locks[lockKey]; ok && time.Since(t) < lockTTL {
			r.mu.Unlock()

			logger.Log.Warnw("callback ignored due to lock",
				"user_id", cb.From.ID,
				"action", action,
			)
			return
		}
		r.locks[lockKey] = time.Now()
		r.mu.Unlock()

		for prefix, h := range r.callbackHandlers {
			if len(cb.Data) >= len(prefix) && cb.Data[:len(prefix)] == prefix {
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

func (r *Router) cleanup() {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		now := time.Now()
		r.mu.Lock()
		for k, t := range r.locks {
			if now.Sub(t) > lockTTL {
				delete(r.locks, k)
			}
		}
		r.mu.Unlock()
	}
}

func buildLockKey(cb *tgbotapi.CallbackQuery, action string) string {
	return fmt.Sprintf(
		"%d:%d:%s",
		cb.Message.Chat.ID,
		cb.Message.MessageID,
		action,
	)
}

func extractPrefix(data string) string {
	for i := 0; i < len(data); i++ {
		if data[i] == ':' {
			return data[:i]
		}
	}
	return data
}
