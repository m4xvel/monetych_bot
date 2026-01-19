package telegram

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func (h *Handler) handleAcceptSelect(
	ctx context.Context,
	cb *tgbotapi.CallbackQuery,
) {
	chatID := cb.Message.Chat.ID

	h.bot.Request(tgbotapi.NewCallback(cb.ID, ""))
	parts := strings.Split(cb.Data, ":")
	if len(parts) < 4 {
		return
	}

	orderID, _ := strconv.Atoi(parts[1])
	messageUserID, _ := strconv.Atoi(parts[2])
	chatUserID, _ := strconv.ParseInt(parts[3], 10, 64)
	expertID, _ := strconv.Atoi(parts[4])

	err := h.orderService.SetAcceptedStatus(ctx, orderID)
	if err != nil {
		return
	}

	order, _ := h.orderService.GetOrderByID(ctx, orderID)
	expert, _ := h.expertService.GetExpertByID(expertID)

	threadID, err := h.createForumTopic(
		h.textDynamic.TitleOrderTopic(
			orderID,
			order.GameNameAtPurchase,
			order.GameTypeNameAtPurchase,
		),
		expert.TopicID,
	)
	if err != nil {
		fmt.Printf("failed to create forum topic: %v", err)
		return
	}

	err = h.orderService.SetExpertData(ctx, orderID, expertID, threadID)
	if err != nil {
		return
	}

	h.deleteOrderMessage(ctx, orderID)

	h.bot.Send(tgbotapi.NewMessage(
		chatID,
		h.textDynamic.AssessorAcceptedOrder(orderID, order.GameNameAtPurchase, order.GameTypeNameAtPurchase),
	))

	h.renderControlPanel(expert.TopicID, threadID, order)

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
