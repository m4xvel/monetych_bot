package telegram

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/m4xvel/monetych_bot/internal/apperr"
	"github.com/m4xvel/monetych_bot/internal/logger"
)

func (h *Handler) handleAcceptSelect(
	ctx context.Context,
	cb *tgbotapi.CallbackQuery,
) {
	chatID := cb.Message.Chat.ID
	h.answerCallback(cb, "")

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

	if err := h.callbackTokenService.Consume(
		ctx,
		tokenCallback,
		"accept",
		&payload,
	); err != nil {
		if isInvalidToken(err) {
			logger.Log.Warnw("invalid accept callback token",
				"chat_id", chatID,
				"data", cb.Data,
				"err", err,
			)
			return
		}
		logger.Log.Errorw("failed to consume accept callback token",
			"chat_id", chatID,
			"err", err,
		)
		return
	}

	chatUserID := payload.ChatID
	messageUserID := payload.UserMessageID
	orderID := payload.OrderID
	expertID := payload.ExpertID

	if err := h.callbackTokenService.DeleteByActionAndOrderID(
		ctx,
		"cancel",
		orderID,
	); err != nil {
		logger.Log.Errorw("failed to delete cancel callbacks",
			"order_id", orderID,
			"err", err,
		)
	}

	if err := h.orderService.SetAcceptedStatus(ctx, orderID); err != nil {
		if isOrderAlreadyProcessed(err) {
			if err := h.callbackTokenService.DeleteByActionAndOrderID(
				ctx,
				"accept",
				orderID,
			); err != nil {
				logger.Log.Errorw("failed to delete accept callbacks",
					"order_id", orderID,
					"err", err,
				)
			}
			logger.Log.Infow("order already processed on accept",
				"order_id", orderID,
				"err", err,
			)
			return
		}
		logger.Log.Warnw("failed to accept order",
			"order_id", orderID,
			"err", err,
		)
		return
	}

	if err := h.callbackTokenService.DeleteByActionAndOrderID(
		ctx,
		"accept",
		orderID,
	); err != nil {
		logger.Log.Errorw("failed to delete accept callbacks",
			"order_id", orderID,
			"err", err,
		)
	}

	logger.Log.Infow("order accepted",
		"order_id", orderID,
		"expert_id", expertID,
	)

	order, err := h.orderService.GetOrderByID(ctx, orderID)
	if err != nil || order == nil {
		logger.Log.Errorw("failed to get order after accept",
			"order_id", orderID,
			"err", err,
		)
		return
	}

	expert, err := h.expertService.GetExpertByID(expertID)
	if err != nil {
		logger.Log.Errorw("failed to get expert",
			"expert_id", expertID,
			"err", err,
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

	if _, err := h.bot.Send(tgbotapi.NewMessage(
		chatID,
		h.textDynamic.AssessorAcceptedOrder(orderID, order.GameNameAtPurchase, order.GameTypeNameAtPurchase),
	)); err != nil {
		wrapped := wrapTelegramErr("telegram.send_expert_accept_notice", err)
		logger.Log.Errorw("failed to send expert accept notice",
			"order_id", orderID,
			"expert_id", expertID,
			"err", wrapped,
		)
	}

	h.renderControlPanel(ctx, expert.TopicID, threadID, order)

	if _, err := h.bot.Request(tgbotapi.NewDeleteMessage(chatUserID, messageUserID)); err != nil {
		wrapped := wrapTelegramErr("telegram.delete_user_message", err)
		logger.Log.Errorw("failed to delete user message",
			"user_chat_id", chatUserID,
			"message_id", messageUserID,
			"err", wrapped,
		)
	}

	message := tgbotapi.NewMessage(
		chatUserID,
		h.textDynamic.AssessorAcceptedYourOrder(order.Token),
	)
	message.ParseMode = "Markdown"

	if _, err := h.bot.Send(message); err != nil {
		wrapped := wrapTelegramErr("telegram.send_accept_to_user", err)
		logger.Log.Errorw("failed to notify user about accepted order",
			"order_id", orderID,
			"user_chat_id", chatUserID,
			"err", wrapped,
		)
	}

	if err := h.stateService.SetStateCommunication(ctx, chatUserID, &orderID); err != nil {
		logger.Log.Errorw("failed to set communication state",
			"order_id", orderID,
			"user_chat_id", chatUserID,
			"err", err,
		)
	}
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
		wrapped := wrapTelegramErr("telegram.create_forum_topic", err)
		logger.Log.Errorw("telegram createForumTopic request failed",
			"topic_id", topicID,
			"err", wrapped,
		)
		return 0, wrapped
	}

	if !resp.Ok {
		wrapped := &apperr.TelegramError{
			Op:      "telegram.create_forum_topic",
			Code:    resp.ErrorCode,
			Message: resp.Description,
		}
		logger.Log.Errorw("telegram createForumTopic api error",
			"description", resp.Description,
			"err", wrapped,
		)
		return 0, wrapped
	}

	var topic struct {
		MessageThreadID int64 `json:"message_thread_id"`
	}
	if err := json.Unmarshal(resp.Result, &topic); err != nil {
		return 0, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return topic.MessageThreadID, nil
}
