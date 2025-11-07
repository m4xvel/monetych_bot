package telegram

import (
	"context"
	"log"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type HandlerFunc func(ctx context.Context, upd tgbotapi.Update)

type CallbackHandlerFunc func(ctx context.Context, cb *tgbotapi.CallbackQuery)

type Router struct {
	commandHandlers  map[string]HandlerFunc
	callbackHandlers map[string]CallbackHandlerFunc
}

func NewRouter() *Router {
	return &Router{
		commandHandlers:  make(map[string]HandlerFunc),
		callbackHandlers: make(map[string]CallbackHandlerFunc),
	}
}

func (r *Router) RegisterCommand(cmd string, handler HandlerFunc) {
	r.commandHandlers[cmd] = handler
}

func (r *Router) RegisterCallback(prefix string, handler CallbackHandlerFunc) {
	r.callbackHandlers[prefix] = handler
}

func (r *Router) Route(ctx context.Context, upd tgbotapi.Update) {
	if upd.Message != nil && upd.Message.IsCommand() {
		if handler, ok := r.commandHandlers[upd.Message.Command()]; ok {
			handler(ctx, upd)
		}
		return
	}
	if upd.CallbackQuery != nil {
		data := upd.CallbackQuery.Data
		for prefix, handler := range r.callbackHandlers {
			if len(data) >= len(prefix) && data[:len(prefix)] == prefix {
				handler(ctx, upd.CallbackQuery)
				return
			}
		}
		log.Printf("no callback handler for %s", data)
	}
}
