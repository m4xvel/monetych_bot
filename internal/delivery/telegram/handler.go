package telegram

import (
	"context"
	"fmt"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/m4xvel/monetych_bot/internal/domain"
	"github.com/m4xvel/monetych_bot/internal/features"
	"github.com/m4xvel/monetych_bot/internal/usecase"
	"github.com/m4xvel/monetych_bot/pkg/utils"
)

type SentOrder struct {
	ChatID    int64
	MessageID int
}

type Handler struct {
	bot                 *tgbotapi.BotAPI
	userService         *usecase.UserService
	stateService        *usecase.StateService
	gameService         *usecase.GameService
	orderService        *usecase.OrderService
	expertService       *usecase.ExpertService
	orderMessageService *usecase.OrderMessageService
	reviewService       *usecase.ReviewService
	router              *Router
	feature             *features.Features
	text                *utils.Messages
	textDynamic         *utils.Dynamic
}

func NewHandler(
	bot *tgbotapi.BotAPI,
	us *usecase.UserService,
	ss *usecase.StateService,
	gs *usecase.GameService,
	os *usecase.OrderService,
	es *usecase.ExpertService,
	rs *usecase.ReviewService,
	oms *usecase.OrderMessageService,
) *Handler {
	h := &Handler{
		bot:                 bot,
		userService:         us,
		stateService:        ss,
		gameService:         gs,
		orderService:        os,
		expertService:       es,
		reviewService:       rs,
		orderMessageService: oms,
		router:              NewRouter(),
		feature:             features.NewFeatures(),
		text:                utils.NewMessages(),
		textDynamic:         utils.NewDynamic(),
	}

	h.registerRoutes()
	return h
}

func (h *Handler) registerRoutes() {
	h.router.RegisterCommand("start", h.handleStartCommand)
	h.router.RegisterCommand("catalog", h.handlerCatalogCommand)

	h.router.RegisterCallback("game:", h.handleGameSelect)
	h.router.RegisterCallback("type:", h.handleTypeSelect)
	h.router.RegisterCallback("order:", h.handleOrderSelect)
	h.router.RegisterCallback("accept:", h.handleAcceptSelect)
	h.router.RegisterCallback("cancel:", h.handlerCancelSelect)
	h.router.RegisterCallback("declined:", h.handleDeclinedSelect)
	h.router.RegisterCallback("declined_reaffirm:",
		h.handleDeclinedReaffirmSelect,
	)
	h.router.RegisterCallback("confirmed:", h.handleConfirmedSelect)
	h.router.RegisterCallback("confirmed_reaffirm:",
		h.handleConfirmedReaffirmSelect,
	)
	h.router.RegisterCallback("back:", h.handleBack)
	h.router.RegisterCallback("accept_client:", h.handleAcceptClientSelect)
	h.router.RegisterCallback("rate:", h.handleRateSelect)

	h.router.RegisterMessageHandler(h.handleMessage)
}

func (h *Handler) Route(ctx context.Context, upd tgbotapi.Update) {
	if !h.stateGuard(ctx, upd) {
		return
	}
	h.router.Route(ctx, upd)
}

func (h *Handler) deleteOrderMessage(ctx context.Context, orderID int) {

	sentOrders, err := h.orderMessageService.GetByOrder(ctx, orderID)
	if err != nil {
		return
	}
	for _, sent := range sentOrders {
		h.bot.Request(tgbotapi.NewDeleteMessage(
			sent.ChatID,
			sent.MessageID,
		))
	}

	h.orderMessageService.MarkDeletedByOrder(ctx, orderID)
}

func (h *Handler) renderControlPanel(
	topicID, threadID int64,
	order *domain.Order) {

	btnAccept := tgbotapi.NewInlineKeyboardButtonData(
		h.text.AcceptText,
		fmt.Sprintf("confirmed:%d:%d:%d", order.ID, topicID, threadID),
	)

	btnDecline := tgbotapi.NewInlineKeyboardButtonData(
		h.text.DeclineText,
		fmt.Sprintf("declined:%d:%d:%d", order.ID, topicID, threadID),
	)

	markup := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(btnAccept),
		tgbotapi.NewInlineKeyboardRow(btnDecline),
	)

	msg := tgbotapi.NewMessage(
		topicID,
		h.textDynamic.ApplicationManagementText(
			order.GameNameAtPurchase,
			order.GameTypeNameAtPurchase,
		),
	)
	msg.MessageThreadID = threadID
	msg.ReplyMarkup = markup

	h.bot.Send(msg)
}

func (h *Handler) renderEditControlPanel(
	messageID int,
	topicID, threadID int64,
	order *domain.Order) {

	btnAccept := tgbotapi.NewInlineKeyboardButtonData(
		h.text.AcceptText,
		fmt.Sprintf("confirmed:%d:%d:%d", order.ID, topicID, threadID),
	)

	btnDecline := tgbotapi.NewInlineKeyboardButtonData(
		h.text.DeclineText,
		fmt.Sprintf("declined:%d:%d:%d", order.ID, topicID, threadID),
	)

	markup := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(btnAccept),
		tgbotapi.NewInlineKeyboardRow(btnDecline),
	)

	editMessage := tgbotapi.NewEditMessageText(
		topicID,
		messageID,
		h.textDynamic.ApplicationManagementText(
			order.GameNameAtPurchase,
			order.GameTypeNameAtPurchase,
		),
	)
	editMessage.ReplyMarkup = &markup

	h.bot.Send(editMessage)
}

func (h *Handler) stateGuard(
	ctx context.Context,
	upd tgbotapi.Update,
) bool {

	chatID, ok := extractChatID(upd)
	if !ok {
		return true
	}

	state, err := h.stateService.GetStateByChatID(ctx, chatID)
	if err != nil {
		return true
	}

	if state.State == domain.StateWritingReview {
		if shouldAutoPublishReview(upd) {
			h.publishPendingReview(ctx, state)
		}
	}

	if state.State != domain.StateCommunication {
		return true
	}

	if upd.Message != nil && upd.Message.IsCommand() {
		h.bot.Send(tgbotapi.NewMessage(
			chatID,
			"Вы уже общаетесь с экспертом.\nИспользуйте чат или дождитесь завершения заказа.",
		))
		return false
	}

	if upd.Message != nil {
		return true
	}

	if upd.CallbackQuery != nil {
		if strings.HasPrefix(upd.CallbackQuery.Data, "accept_client:") {
			return true
		}

		h.answerCallback(
			upd.CallbackQuery,
			"Эта кнопка недоступна во время общения с экспертом",
		)
		return false
	}

	return true
}

func shouldAutoPublishReview(upd tgbotapi.Update) bool {
	if upd.Message != nil {

		if upd.Message.IsCommand() {
			return true
		}

		if upd.Message.Text != "" || upd.Message.Caption != "" {
			return false
		}
	}

	return true
}

func (h *Handler) publishPendingReview(
	ctx context.Context,
	state *domain.UserState,
) {
	if state.ReviewID == nil {
		return
	}

	if err := h.reviewService.Publish(ctx, *state.ReviewID); err != nil {
		return
	}

	h.stateService.SetStateIdle(ctx, *state.UserChatID)
}

func extractChatID(upd tgbotapi.Update) (int64, bool) {
	if upd.Message != nil {
		return upd.Message.Chat.ID, true
	}

	if upd.CallbackQuery != nil {
		return upd.CallbackQuery.Message.Chat.ID, true
	}

	return 0, false
}

func (h *Handler) answerCallback(
	cb *tgbotapi.CallbackQuery,
	text string,
) {
	h.bot.Request(tgbotapi.NewCallback(cb.ID, text))
}
