package telegram

import (
	"context"
	"fmt"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func (h *Handler) handlerCatalogCommand(
	ctx context.Context,
	msg *tgbotapi.Message,
) {
	chatID := msg.Chat.ID
	games, _ := h.gameService.GetAllGames()
	var rows [][]tgbotapi.InlineKeyboardButton
	for _, g := range games {
		btn := tgbotapi.NewInlineKeyboardButtonData(
			g.Name,
			fmt.Sprintf("game:%d", g.ID),
		)
		rows = append(rows, tgbotapi.NewInlineKeyboardRow(btn))
	}

	message := tgbotapi.NewMessage(chatID, h.text.ChooseGame)
	message.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(rows...)
	h.bot.Send(message)

	h.stateService.SetStateIdle(ctx, chatID)
}
