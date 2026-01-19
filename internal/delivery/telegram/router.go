package telegram

import (
	"context"
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

	mu             sync.Mutex
	lockedMessages map[int]time.Time
}

func NewRouter() *Router {
	r := &Router{
		commandHandlers:  make(map[string]HandlerFunc),
		callbackHandlers: make(map[string]CallbackHandlerFunc),
		lockedMessages:   make(map[int]time.Time),
	}

	go r.cleanupLockedMessages()

	return r
}

const lockTTL = 10 * time.Second

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

		msgID := cb.Message.MessageID

		r.mu.Lock()

		if t, ok := r.lockedMessages[msgID]; ok {
			if time.Since(t) < lockTTL {
				r.mu.Unlock()
				return
			}
		}

		r.lockedMessages[msgID] = time.Now()
		r.mu.Unlock()

		data := cb.Data
		for prefix, handler := range r.callbackHandlers {
			if len(data) >= len(prefix) && data[:len(prefix)] == prefix {
				handler(ctx, cb)
				return
			}
		}
		log.Printf("no callback handler for %s", data)
	}

	go func() {
		ticker := time.NewTicker(1 * time.Minute)
		defer ticker.Stop()

		for range ticker.C {
			now := time.Now()
			r.mu.Lock()
			for id, t := range r.lockedMessages {
				if now.Sub(t) > lockTTL {
					delete(r.lockedMessages, id)
				}
			}
			r.mu.Unlock()
		}
	}()
}

func (r *Router) cleanupLockedMessages() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		now := time.Now()
		r.mu.Lock()
		for id, t := range r.lockedMessages {
			if now.Sub(t) > lockTTL {
				delete(r.lockedMessages, id)
			}
		}
		r.mu.Unlock()
	}
}
