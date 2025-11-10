package telegram

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/m4xvel/monetych_bot/internal/usecase"
)

type Handler struct {
	bot             *tgbotapi.BotAPI
	gameService     *usecase.GameService
	userService     *usecase.UserService
	orderService    *usecase.OrderService
	assessorService *usecase.AssessorService
	router          *Router

	mu                  sync.Mutex
	lastProcessedChatID map[int64]int
}

func NewHandler(
	bot *tgbotapi.BotAPI,
	gs *usecase.GameService,
	us *usecase.UserService,
	os *usecase.OrderService,
	as *usecase.AssessorService) *Handler {
	h := &Handler{
		bot:                 bot,
		gameService:         gs,
		userService:         us,
		orderService:        os,
		assessorService:     as,
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
	h.router.RegisterCallback("order:", h.handleOrderSelect)
	h.router.RegisterCallback("accept:", h.handleAcceptSelect)
}

func (h *Handler) Route(ctx context.Context, upd tgbotapi.Update) {
	h.router.Route(ctx, upd)
}

func (h *Handler) showInlineKeyboardVerification(chatID int64, text string, isVerifyAPI bool, nameGame, nameType string) {
	msg := tgbotapi.NewMessage(chatID, text)
	verificationButton := tgbotapi.NewInlineKeyboardButtonData("Пройти верификацию", fmt.Sprintf("verify:%t:%s:%s", isVerifyAPI, nameGame, nameType)) // Сюда Callback с API
	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(verificationButton),
	)
	msg.ReplyMarkup = keyboard
	h.bot.Send(msg)
}

func (h *Handler) contactAnAppraiser(chatID int64, nameGame, nameType string) {
	msg := tgbotapi.NewMessage(chatID, "Свяжитесь с оценщиком 📩")
	verificationButton := tgbotapi.NewInlineKeyboardButtonData("Связаться 💬", fmt.Sprintf("order:%s:%s", nameGame, nameType))
	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(verificationButton),
	)
	msg.ReplyMarkup = keyboard
	h.bot.Send(msg)
}

func (h *Handler) createForumTopic(
	ctx context.Context,
	topicName string,
	assessorID int64,
) (int64, error) {
	chatID := h.assessorService.GetTopicIDByTgID(ctx, assessorID)
	params := tgbotapi.Params{
		"chat_id": fmt.Sprint(chatID),
		"name":    topicName,
	}

	resp, err := h.bot.MakeRequest("createForumTopic", params)
	if err != nil {
		return 0, fmt.Errorf("createForumTopic failed: %w", err)
	}
	var result struct {
		Ok     bool `json:"ok"`
		Result struct {
			MessageThreadID int64 `json:"message_thread_id"`
		} `json:"result"`
	}

	if err := json.Unmarshal(resp.Result, &result); err != nil {
		return 0, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if !result.Ok {
		return 0, fmt.Errorf("telegram API returned not ok")
	}

	return result.Result.MessageThreadID, nil
}
