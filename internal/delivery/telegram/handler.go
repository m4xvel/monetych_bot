package telegram

import (
	"context"
	"sync"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/m4xvel/monetych_bot/internal/usecase"
)

type Handler struct {
	bot         *tgbotapi.BotAPI
	gameService *usecase.GameService
	router      *Router

	mu                  sync.Mutex
	lastProcessedChatID map[int64]int
}

func NewHandler(bot *tgbotapi.BotAPI, gs *usecase.GameService) *Handler {
	h := &Handler{
		bot:                 bot,
		gameService:         gs,
		router:              NewRouter(),
		lastProcessedChatID: make(map[int64]int),
	}

	h.registerRoutes()
	return h
}

func (h *Handler) shouldProcess(chatID int64, messageID int) bool {
	h.mu.Lock()
	defer h.mu.Unlock()

	if lastID, ok := h.lastProcessedChatID[chatID]; ok && lastID == messageID {
		return false
	}

	h.lastProcessedChatID[chatID] = messageID
	return true
}

func (h *Handler) registerRoutes() {
	h.router.RegisterCommand("start", h.handleStartCommand)

	h.router.RegisterCallback("game:", h.handleGameSelect)
	h.router.RegisterCallback("type:", h.handleTypeSelect)
}

func (h *Handler) Route(ctx context.Context, upd tgbotapi.Update) {
	h.router.Route(ctx, upd)
}
