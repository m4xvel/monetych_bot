package telegram

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
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
			if handler, ok := r.commandHandlers[msg.Command()]; ok {
				handler(ctx, msg)
				return
			}
		}

		if r.messageHandler != nil {
			r.messageHandler(ctx, msg)
			return
		}

	case upd.CallbackQuery != nil:
		cb := upd.CallbackQuery

		action := extractPrefix(cb.Data)
		lockKey := buildLockKey(cb, action)

		r.mu.Lock()
		if t, ok := r.locks[lockKey]; ok && time.Since(t) < lockTTL {
			r.mu.Unlock()
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

		log.Printf("no callback handler for %s", cb.Data)
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
