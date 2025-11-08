package telegram

import (
	"context"
	"fmt"
	"sync"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/m4xvel/monetych_bot/internal/usecase"
)

type Handler struct {
	bot          *tgbotapi.BotAPI
	gameService  *usecase.GameService
	userService  *usecase.UserService
	orderService *usecase.OrderService
	router       *Router

	mu                  sync.Mutex
	lastProcessedChatID map[int64]int
}

func NewHandler(
	bot *tgbotapi.BotAPI,
	gs *usecase.GameService,
	us *usecase.UserService,
	os *usecase.OrderService) *Handler {
	h := &Handler{
		bot:                 bot,
		gameService:         gs,
		userService:         us,
		orderService:        os,
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
	h.router.RegisterCallback("verify:", h.handleVerifySelect)
}

func (h *Handler) Route(ctx context.Context, upd tgbotapi.Update) {
	h.router.Route(ctx, upd)
}

func (h *Handler) showInlineKeyboardVerification(chatID int64, text string, isVerifyAPI bool) {
	msg := tgbotapi.NewMessage(chatID, text)
	verificationButton := tgbotapi.NewInlineKeyboardButtonData("Пройти верификацию", fmt.Sprintf("verify:%t", isVerifyAPI)) // Сюда Callback с API
	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(verificationButton),
	)
	msg.ReplyMarkup = keyboard
	h.bot.Send(msg)
}

func (h *Handler) contactAnAppraiser(chatID int64) {
	msg := tgbotapi.NewMessage(chatID, "Свяжитесь с оценщиком 📩")
	verificationButton := tgbotapi.NewInlineKeyboardButtonData("Связаться 💬", "order:")
	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(verificationButton),
	)
	msg.ReplyMarkup = keyboard
	h.bot.Send(msg)
}
