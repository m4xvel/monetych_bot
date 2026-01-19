package telegram

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func (h *Handler) handleOrderSelect(
	ctx context.Context,
	cb *tgbotapi.CallbackQuery,
) {
	chatID := cb.Message.Chat.ID
	messageID := cb.Message.MessageID

	h.bot.Request(tgbotapi.NewCallback(cb.ID, ""))

	parts := strings.Split(cb.Data, ":")
	if len(parts) < 3 {
		return
	}
	gameID, _ := strconv.Atoi(parts[1])
	gameTypeID, _ := strconv.Atoi(parts[2])

	u, _ := h.userService.GetByChatID(ctx, chatID)
	g, _ := h.gameService.GetGameByID(gameID)
	t, _ := h.gameService.GetTypeByID(gameTypeID)

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
		return
	}

	if id == 0 {
		h.bot.Send(tgbotapi.NewEditMessageText(
			chatID,
			messageID,
			h.text.AlreadyActiveOrder,
		))
		return
	}

	message := tgbotapi.NewMessage(
		chatID,
		h.text.WaitingAssessor,
	)

	delete := tgbotapi.NewDeleteMessage(chatID, messageID)

	h.bot.Send(delete)

	btn := tgbotapi.NewInlineKeyboardButtonData(
		h.text.DeclineText,
		fmt.Sprintf("cancel:%d", id),
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
	experts, _ := h.expertService.GetAllExperts()
	for _, e := range experts {
		btn := tgbotapi.NewInlineKeyboardButtonData(
			"Принять",
			fmt.Sprintf("accept:%d:%d:%d:%d", orderID, messageID, chatID, e.ID),
		)

		message := tgbotapi.NewMessage(
			e.ChatID,
			h.textDynamic.NewOrder(orderID, gameName, gameTypeName),
		)

		message.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(btn),
		)

		send, _ := h.bot.Send(message)

		h.orderMessageService.Save(ctx, orderID, send.Chat.ID, send.MessageID)
	}
}
