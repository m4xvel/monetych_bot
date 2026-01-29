package telegram

import (
	"context"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/m4xvel/monetych_bot/internal/domain"
	"github.com/m4xvel/monetych_bot/internal/features"
	"github.com/m4xvel/monetych_bot/internal/logger"
	"github.com/m4xvel/monetych_bot/internal/usecase"
	"github.com/m4xvel/monetych_bot/pkg/utils"
)

type SentOrder struct {
	ChatID    int64 `json:"chat_id"`
	MessageID int   `json:"message_id"`
}

type GameSelectPayload struct {
	ChatID int64 `json:"chat_id"`
	GameID int   `json:"game_id"`
}

type TypeSelectPayload struct {
	ChatID int64 `json:"chat_id"`
	GameID int   `json:"game_id"`
	TypeID int   `json:"type_id"`
}

type OrderSelectPayload struct {
	ChatID int64 `json:"chat_id"`
	GameID int   `json:"game_id"`
	TypeID int   `json:"type_id"`
}

type CancelOrderSelectPayload struct {
	ChatID  int64 `json:"chat_id"`
	OrderID int   `json:"order_id"`
}

type AcceptOrderSelectPayload struct {
	ChatID        int64 `json:"chat_id"`
	OrderID       int   `json:"order_id"`
	UserMessageID int   `json:"user_message_id"`
	ExpertID      int   `json:"expert_id"`
}

type ConfirmedAndDeclinedOrderSelectPayload struct {
	OrderID  int   `json:"order_id"`
	TopicID  int64 `json:"topic_id"`
	ThreadID int64 `json:"thread_id"`
}

type RateSelectPayload struct {
	ChatID  int64 `json:"chat_id"`
	Rate    int   `json:"rate"`
	OrderID int   `json:"order_id"`
}

type SearchPayload struct {
	ChatID  int64 `json:"chat_id"`
	OrderID int   `json:"order_id"`
}

type Handler struct {
	bot                     *tgbotapi.BotAPI
	userService             *usecase.UserService
	stateService            *usecase.StateService
	gameService             *usecase.GameService
	orderService            *usecase.OrderService
	expertService           *usecase.ExpertService
	supportService          *usecase.SupportService
	orderMessageService     *usecase.OrderMessageService
	orderChatMessageService *usecase.OrderChatMessageService
	reviewService           *usecase.ReviewService
	callbackTokenService    *usecase.CallbackTokenService
	router                  *Router
	feature                 *features.Features
	text                    *utils.Messages
	textDynamic             *utils.Dynamic
}

func NewHandler(
	bot *tgbotapi.BotAPI,
	us *usecase.UserService,
	ss *usecase.StateService,
	gs *usecase.GameService,
	os *usecase.OrderService,
	es *usecase.ExpertService,
	sups *usecase.SupportService,
	rs *usecase.ReviewService,
	oms *usecase.OrderMessageService,
	ocms *usecase.OrderChatMessageService,
	cts *usecase.CallbackTokenService,
) *Handler {
	h := &Handler{
		bot:                     bot,
		userService:             us,
		stateService:            ss,
		gameService:             gs,
		orderService:            os,
		expertService:           es,
		supportService:          sups,
		reviewService:           rs,
		orderMessageService:     oms,
		orderChatMessageService: ocms,
		callbackTokenService:    cts,
		router:                  NewRouter(),
		feature:                 features.NewFeatures(),
		text:                    utils.NewMessages(),
		textDynamic:             utils.NewDynamic(),
	}

	h.registerRoutes()
	return h
}

func (h *Handler) registerRoutes() {
	h.router.RegisterCommand("start", h.handleStartCommand)
	h.router.RegisterCommand("catalog", h.handlerCatalogCommand)
	h.router.RegisterCommand("support", h.handlerSupportCommand)
	h.router.RegisterCommand("search", h.supportOnly(h.SearchCommand))

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

	h.router.RegisterCallback("show_media:", h.handleShowMedia)

	h.router.RegisterMessageHandler(h.handleMessage)
}

func (h *Handler) Route(ctx context.Context, upd tgbotapi.Update) {
	if !h.expertGuard(upd) {
		logger.Log.Warnw("update blocked by expert guard")
		return
	}
	if !h.supportGuard(upd) {
		logger.Log.Warnw("update blocked by support guard")
		return
	}
	if !h.stateGuard(ctx, upd) {
		logger.Log.Warnw("update blocked by state guard")
		return
	}
	h.router.Route(ctx, upd)
}

func (h *Handler) deleteOrderMessage(ctx context.Context, orderID int) {
	sentOrders, err := h.orderMessageService.GetByOrder(ctx, orderID)
	if err != nil {
		logger.Log.Errorw("failed to get order messages",
			"order_id", orderID,
			"err", err,
		)
		return
	}

	logger.Log.Infow("deleting order messages",
		"order_id", orderID,
		"count", len(sentOrders),
	)

	for _, sent := range sentOrders {
		h.bot.Request(tgbotapi.NewDeleteMessage(
			sent.ChatID,
			sent.MessageID,
		))
	}

	h.orderMessageService.MarkDeletedByOrder(ctx, orderID)
}

func (h *Handler) renderControlPanel(
	ctx context.Context,
	topicID, threadID int64,
	order *domain.Order) {

	tokenConfirmed, err := h.callbackTokenService.Create(
		ctx,
		"confirmed",
		&ConfirmedAndDeclinedOrderSelectPayload{
			OrderID:  order.ID,
			TopicID:  topicID,
			ThreadID: threadID,
		},
	)
	if err != nil {
		logger.Log.Errorw("failed to create confirmed order callback token",
			"err", err,
		)
	}

	btnAccept := tgbotapi.NewInlineKeyboardButtonData(
		h.text.AcceptText,
		"confirmed:"+tokenConfirmed,
	)

	tokenDeclined, err := h.callbackTokenService.Create(
		ctx,
		"declined",
		&ConfirmedAndDeclinedOrderSelectPayload{
			OrderID:  order.ID,
			TopicID:  topicID,
			ThreadID: threadID,
		},
	)
	if err != nil {
		logger.Log.Errorw("failed to create declined order callback token",
			"err", err,
		)
	}

	btnDecline := tgbotapi.NewInlineKeyboardButtonData(
		h.text.DeclineText,
		"declined:"+tokenDeclined,
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
	ctx context.Context,
	messageID int,
	topicID, threadID int64,
	order *domain.Order) {

	tokenConfirmed, err := h.callbackTokenService.Create(
		ctx,
		"confirmed",
		&ConfirmedAndDeclinedOrderSelectPayload{
			OrderID:  order.ID,
			TopicID:  topicID,
			ThreadID: threadID,
		},
	)
	if err != nil {
		logger.Log.Errorw("failed to create confirmed order callback token",
			"err", err,
		)
	}

	btnAccept := tgbotapi.NewInlineKeyboardButtonData(
		h.text.AcceptText,
		"confirmed:"+tokenConfirmed,
	)

	tokenDeclined, err := h.callbackTokenService.Create(
		ctx,
		"declined",
		&ConfirmedAndDeclinedOrderSelectPayload{
			OrderID:  order.ID,
			TopicID:  topicID,
			ThreadID: threadID,
		},
	)
	if err != nil {
		logger.Log.Errorw("failed to create declined order callback token",
			"err", err,
		)
	}

	btnDecline := tgbotapi.NewInlineKeyboardButtonData(
		h.text.DeclineText,
		"declined:"+tokenDeclined,
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

func (h *Handler) expertGuard(
	upd tgbotapi.Update,
) bool {

	chatID, ok := extractChatID(upd)
	if !ok {
		return true
	}

	experts, err := h.expertService.GetAllExperts()
	if err != nil {
		return true
	}

	isExpert := false
	for _, e := range experts {
		if chatID == e.TopicID {
			isExpert = true
			break
		}
	}

	if !isExpert {
		return true
	}

	if upd.CallbackQuery != nil {
		return true
	}

	if upd.Message != nil && !upd.Message.IsCommand() {
		return true
	}

	logger.Log.Infow("expert interaction detected",
		"chat_id", chatID,
	)

	if upd.Message != nil && upd.Message.IsCommand() &&
		upd.Message.Command() == "start" {
		return true
	}

	logger.Log.Warnw("expert action blocked",
		"chat_id", chatID,
	)

	return false
}

func (h *Handler) supportGuard(
	upd tgbotapi.Update,
) bool {

	chatID, ok := extractChatID(upd)
	if !ok {
		return true
	}

	support := h.supportService.GetSupport()

	if chatID != support.ChatID {
		return true
	}

	if upd.Message != nil && upd.Message.IsCommand() {
		switch upd.Message.Command() {
		case "start", "search":
			return true
		default:
			logger.Log.Warnw("support action blocked",
				"chat_id", chatID,
			)
			return false
		}
	}

	if upd.Message != nil {
		logger.Log.Warnw("support action blocked",
			"chat_id", chatID,
		)
		return false
	}

	return true
}

func (h *Handler) stateGuard(
	ctx context.Context,
	upd tgbotapi.Update,
) bool {
	if upd.Message != nil {
		if upd.Message.Chat.IsSuperGroup() {
			return true
		}
	}

	if upd.CallbackQuery != nil {
		if upd.CallbackQuery.Message.Chat.IsSuperGroup() {
			return true
		}
	}

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
			logger.Log.Infow("auto publishing pending review",
				"user_chat_id", *state.UserChatID,
				"review_id", *state.ReviewID,
			)
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

		logger.Log.Warnw("command blocked during communication",
			"user_chat_id", chatID,
			"state", state.State,
		)

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

		logger.Log.Warnw("callback blocked during communication",
			"user_chat_id", chatID,
			"data", upd.CallbackQuery.Data,
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
		logger.Log.Errorw("failed to publish review",
			"review_id", *state.ReviewID,
			"err", err,
		)
		return
	}

	logger.Log.Infow("review published",
		"review_id", *state.ReviewID,
	)

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

func (h *Handler) supportOnly(handler HandlerFunc) HandlerFunc {
	return func(ctx context.Context, msg *tgbotapi.Message) {
		support := h.supportService.GetSupport()
		if msg.Chat.ID != support.ChatID {
			return
		}
		handler(ctx, msg)
	}
}
