package telegram

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/m4xvel/monetych_bot/internal/usecase"
	"github.com/m4xvel/monetych_bot/pkg/utils"
)

type SentOrder struct {
	ChatID    int64
	MessageID int
}

type Handler struct {
	bot             *tgbotapi.BotAPI
	gameService     *usecase.GameService
	userService     *usecase.UserService
	orderService    *usecase.OrderService
	assessorService *usecase.AssessorService
	stateService    *usecase.StateService
	reviewService   *usecase.ReviewService
	router          *Router
	text            *utils.Messages
	textDynamic     *utils.Dynamic

	mu                  sync.Mutex
	lastProcessedChatID map[int64]int
	orderMessages       map[int][]SentOrder
}

func NewHandler(
	bot *tgbotapi.BotAPI,
	gs *usecase.GameService,
	us *usecase.UserService,
	os *usecase.OrderService,
	as *usecase.AssessorService,
	ss *usecase.StateService,
	rs *usecase.ReviewService,
) *Handler {
	h := &Handler{
		bot:                 bot,
		gameService:         gs,
		userService:         us,
		orderService:        os,
		assessorService:     as,
		stateService:        ss,
		reviewService:       rs,
		text:                utils.NewMessages(),
		textDynamic:         utils.NewDynamic(),
		router:              NewRouter(),
		lastProcessedChatID: make(map[int64]int),
		orderMessages:       make(map[int][]SentOrder),
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

func (h *Handler) addSentOrder(orderID int, sent SentOrder) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.orderMessages[orderID] = append(h.orderMessages[orderID], sent)
}

func (h *Handler) deleteSentOrders(orderID int) {
	h.mu.Lock()
	defer h.mu.Unlock()

	sentOrders, ok := h.orderMessages[orderID]
	if !ok {
		return
	}
	for _, sent := range sentOrders {
		h.bot.Request(tgbotapi.NewDeleteMessage(sent.ChatID, sent.MessageID))
	}
	delete(h.orderMessages, orderID)
}

func (h *Handler) registerRoutes() {
	h.router.RegisterCommand("start", h.handleStartCommand)
	h.router.RegisterCommand("catalog", h.handleCatalogCommand)
	h.router.RegisterCommand("support", h.handleSupportCommand)

	h.router.RegisterCallback("game:", h.handleGameSelect)
	h.router.RegisterCallback("type:", h.handleTypeSelect)
	h.router.RegisterCallback("verify:", h.handleVerifySelect)
	h.router.RegisterCallback("order:", h.handleOrderSelect)
	h.router.RegisterCallback("accept:", h.handleAcceptSelect)
	h.router.RegisterCallback("order_accept:", h.handleOrderAcceptAssessor)
	h.router.RegisterCallback("order_decline:", h.handleOrderDeclineAssessor)
	h.router.RegisterCallback("order_accept_client:", h.handleOrderAcceptClient)
	h.router.RegisterCallback("rate:", h.handleRateSelect)

	h.router.RegisterMessageHandler(h.handleMessage)
}

func (h *Handler) Route(ctx context.Context, upd tgbotapi.Update) {
	h.router.Route(ctx, upd)
}

func (h *Handler) showInlineKeyboardVerification(chatID int64, text string, isVerifyAPI bool, nameGame, nameType string) {
	msg := tgbotapi.NewMessage(chatID, text)
	verificationButton := tgbotapi.NewInlineKeyboardButtonData(h.text.VerifyButtonText, fmt.Sprintf("verify:%t:%s:%s", isVerifyAPI, nameGame, nameType)) // Сюда Callback с API
	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(verificationButton),
	)
	msg.ReplyMarkup = keyboard
	h.bot.Send(msg)
}

func (h *Handler) contactAnAppraiser(chatID int64, nameGame, nameType string) {
	msg := tgbotapi.NewMessage(chatID, h.text.ContactAppraiserText)
	verificationButton := tgbotapi.NewInlineKeyboardButtonData(h.text.ContactText, fmt.Sprintf("order:%s:%s", nameGame, nameType))
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
	assessor, err := h.assessorService.GetByTgID(ctx, assessorID)
	if err != nil {
		return 0, fmt.Errorf("get assessor by tg id: %w", err)
	}
	params := tgbotapi.Params{
		"chat_id": fmt.Sprint(assessor.TopicID),
		"name":    topicName,
	}

	resp, err := h.bot.MakeRequest("createForumTopic", params)
	if err != nil {
		return 0, fmt.Errorf("createForumTopic failed: %w", err)
	}

	if !resp.Ok {
		return 0, fmt.Errorf("telegram api error: %s", resp.Description)
	}

	var topic struct {
		MessageThreadID int64 `json:"message_thread_id"`
	}
	if err := json.Unmarshal(resp.Result, &topic); err != nil {
		return 0, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return topic.MessageThreadID, nil
}
