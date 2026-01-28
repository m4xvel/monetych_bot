package telegram

import (
	"context"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/m4xvel/monetych_bot/internal/logger"
)

func (h *Handler) handleOrderSelect(
	ctx context.Context,
	cb *tgbotapi.CallbackQuery,
) {
	chatID := cb.Message.Chat.ID
	messageID := cb.Message.MessageID
	h.bot.Request(tgbotapi.NewCallback(cb.ID, ""))

	logger.Log.Infow("order creation initiated",
		"chat_id", chatID,
	)

	parts := strings.Split(cb.Data, ":")
	if len(parts) != 2 {
		logger.Log.Warnw("invalid order callback data",
			"chat_id", chatID,
			"data", cb.Data,
		)
		return
	}

	tokenCallback := parts[1]

	var payload OrderSelectPayload

	h.callbackTokenService.Consume(
		ctx,
		tokenCallback,
		"order",
		&payload,
	)

	if payload.ChatID != cb.From.ID {
		return
	}

	gameID := payload.GameID
	gameTypeID := payload.TypeID

	u, err := h.userService.GetByChatID(ctx, chatID)
	if err != nil || u == nil {
		logger.Log.Warnw("user not found for order creation",
			"chat_id", chatID,
		)
		return
	}

	g, err := h.gameService.GetGameByID(gameID)
	if err != nil {
		logger.Log.Warnw("game not found for order creation",
			"chat_id", chatID,
			"game_id", gameID,
		)
		return
	}

	t, err := h.gameService.GetTypeByID(gameTypeID)
	if err != nil {
		logger.Log.Warnw("game type not found for order creation",
			"chat_id", chatID,
			"game_type_id", gameTypeID,
		)
		return
	}

	id, err := h.orderService.CreateOrder(
		ctx,
		u.ID,
		gameID,
		gameTypeID,
		u.Name,
		g.Name,
		t.Name,
	)
	if err != nil {
		logger.Log.Errorw("failed to create order",
			"chat_id", chatID,
			"user_id", u.ID,
			"game_id", gameID,
			"game_type_id", gameTypeID,
			"err", err,
		)
		return
	}

	if id == 0 {
		logger.Log.Warnw("order already active",
			"chat_id", chatID,
			"user_id", u.ID,
			"game_id", gameID,
			"game_type_id", gameTypeID,
		)

		h.bot.Send(tgbotapi.NewEditMessageText(
			chatID,
			messageID,
			h.text.AlreadyActiveOrder,
		))
		return
	}

	logger.Log.Infow("order created via telegram",
		"order_id", id,
		"user_id", u.ID,
		"chat_id", chatID,
	)

	message := tgbotapi.NewMessage(
		chatID,
		h.text.WaitingAssessor,
	)

	delete := tgbotapi.NewDeleteMessage(chatID, messageID)

	h.bot.Send(delete)

	token, err := h.callbackTokenService.Create(
		ctx,
		"cancel",
		&CancelOrderSelectPayload{
			ChatID:  chatID,
			OrderID: id,
		},
	)
	if err != nil {
		logger.Log.Errorw("failed to create cancel order callback token",
			"chat_id", chatID,
			"err", err,
		)
	}

	btn := tgbotapi.NewInlineKeyboardButtonData(
		h.text.DeclineText,
		"cancel:"+token,
	)

	message.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(btn),
	)

	send, _ := h.bot.Send(message)

	h.notifyExpertsAboutOrder(ctx, id, send.MessageID, chatID, g.Name, t.Name)
}

func (h *Handler) notifyExpertsAboutOrder(
	ctx context.Context,
	orderID, messageID int,
	chatID int64,
	gameName, gameTypeName string,
) {
	logger.Log.Infow("notifying experts about order",
		"order_id", orderID,
	)

	experts, err := h.expertService.GetAllExperts()
	if err != nil {
		logger.Log.Errorw("failed to get experts for order notification",
			"order_id", orderID,
			"err", err,
		)
		return
	}

	for _, e := range experts {
		token, err := h.callbackTokenService.Create(
			ctx,
			"accept",
			&AcceptOrderSelectPayload{
				ChatID:        chatID,
				OrderID:       orderID,
				UserMessageID: messageID,
				ExpertID:      e.ID,
			},
		)
		if err != nil {
			logger.Log.Errorw("failed to create accept order callback token",
				"chat_id", chatID,
				"err", err,
			)
		}

		btn := tgbotapi.NewInlineKeyboardButtonData(
			"Принять",
			"accept:"+token,
		)

		message := tgbotapi.NewMessage(
			e.ChatID,
			h.textDynamic.NewOrder(orderID, gameName, gameTypeName),
		)

		message.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(btn),
		)

		send, err := h.bot.Send(message)
		if err != nil {
			logger.Log.Errorw("failed to notify expert",
				"order_id", orderID,
				"expert_id", e.ID,
				"err", err,
			)
			continue
		}

		if err := h.orderMessageService.Save(
			ctx,
			orderID,
			send.Chat.ID,
			send.MessageID,
		); err != nil {
			logger.Log.Errorw("failed to save order message",
				"order_id", orderID,
				"expert_id", e.ID,
				"err", err,
			)
		}

		logger.Log.Infow("experts notified",
			"order_id", orderID,
			"experts_count", len(experts),
		)
	}
}
