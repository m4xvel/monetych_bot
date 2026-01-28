package telegram

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/m4xvel/monetych_bot/internal/logger"
)

func (h *Handler) handleAcceptSelect(
	ctx context.Context,
	cb *tgbotapi.CallbackQuery,
) {
	chatID := cb.Message.Chat.ID
	h.bot.Request(tgbotapi.NewCallback(cb.ID, ""))

	logger.Log.Infow("expert accepted order click",
		"callback_data", cb.Data,
		"expert_chat_id", chatID,
	)

	parts := strings.Split(cb.Data, ":")
	if len(parts) != 2 {
		logger.Log.Warnw("invalid accept callback data",
			"chat_id", chatID,
			"data", cb.Data,
		)
		return
	}

	tokenCallback := parts[1]

	var payload AcceptOrderSelectPayload

	h.callbackTokenService.Consume(
		ctx,
		tokenCallback,
		"accept",
		&payload,
	)

	chatUserID := payload.ChatID
	messageUserID := payload.UserMessageID
	orderID := payload.OrderID
	expertID := payload.ExpertID

	h.callbackTokenService.DeleteByActionAndOrderID(
		ctx,
		"cancel",
		orderID,
	)

	if err := h.orderService.SetAcceptedStatus(ctx, orderID); err != nil {
		logger.Log.Warnw("failed to accept order",
			"order_id", orderID,
			"err", err,
		)
		return
	}

	logger.Log.Infow("order accepted",
		"order_id", orderID,
		"expert_id", expertID,
	)

	order, err := h.orderService.GetOrderByID(ctx, orderID)
	if err != nil || order == nil {
		logger.Log.Errorw("failed to get order after accept",
			"order_id", orderID,
		)
		return
	}

	expert, err := h.expertService.GetExpertByID(expertID)
	if err != nil {
		logger.Log.Errorw("failed to get expert",
			"expert_id", expertID,
		)
		return
	}

	threadID, err := h.createForumTopic(
		h.textDynamic.TitleOrderTopic(
			orderID,
			order.GameNameAtPurchase,
			order.GameTypeNameAtPurchase,
		),
		expert.TopicID,
	)
	if err != nil {
		logger.Log.Errorw("failed to create forum topic",
			"order_id", orderID,
			"expert_id", expertID,
			"err", err,
		)
		return
	}

	logger.Log.Infow("forum topic created",
		"order_id", orderID,
		"thread_id", threadID,
	)

	if err := h.orderService.SetExpertData(ctx, orderID, expertID, threadID); err != nil {
		logger.Log.Errorw("failed to assign expert to order",
			"order_id", orderID,
			"expert_id", expertID,
			"err", err,
		)
		return
	}

	h.deleteOrderMessage(ctx, orderID)

	h.bot.Send(tgbotapi.NewMessage(
		chatID,
		h.textDynamic.AssessorAcceptedOrder(orderID, order.GameNameAtPurchase, order.GameTypeNameAtPurchase),
	))

	h.renderControlPanel(ctx, expert.TopicID, threadID, order)

	h.bot.Send(tgbotapi.NewDeleteMessage(chatUserID, messageUserID))

	message := tgbotapi.NewMessage(
		chatUserID,
		h.textDynamic.AssessorAcceptedYourOrder(order.Token),
	)
	message.ParseMode = "Markdown"

	h.bot.Send(message)

	h.stateService.SetStateCommunication(ctx, chatUserID, &orderID)
}

func (h *Handler) createForumTopic(
	topicName string,
	topicID int64,
) (int64, error) {

	params := tgbotapi.Params{
		"chat_id": fmt.Sprint(topicID),
		"name":    topicName,
	}

	resp, err := h.bot.MakeRequest("createForumTopic", params)
	if err != nil {
		logger.Log.Errorw("telegram createForumTopic request failed",
			"topic_id", topicID,
			"err", err,
		)
		return 0, err
	}

	if !resp.Ok {
		logger.Log.Errorw("telegram createForumTopic api error",
			"description", resp.Description,
		)
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
