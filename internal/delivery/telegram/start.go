package telegram

import (
	"context"
	"fmt"
	"log"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func (h *Handler) handleStartCommand(ctx context.Context, upd tgbotapi.Update) {
	chatID := upd.Message.Chat.ID

	err := h.userService.AddUserIfNotExists(ctx, chatID)
	if err != nil {
		h.bot.Send(tgbotapi.NewMessage(chatID, "❌ Ошибка добавления в базу данных."))
		return
	}

	games, err := h.gameService.ListGames(ctx)
	if err != nil {
		h.bot.Send(tgbotapi.NewMessage(chatID, "❌ Не удалось загрузить список игр."))
		log.Println(err)
		return
	}

	if len(games) == 0 {
		h.bot.Send(tgbotapi.NewMessage(chatID, "😕 Пока нет доступных игр."))
		return
	}

	var rows [][]tgbotapi.InlineKeyboardButton
	for _, g := range games {
		btn := tgbotapi.NewInlineKeyboardButtonData(g.Name, fmt.Sprintf("game:%d:%s", g.ID, g.Name))
		rows = append(rows, tgbotapi.NewInlineKeyboardRow(btn))
	}

	msg := tgbotapi.NewMessage(chatID, "Выберите игру 🎮")
	msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(rows...)
	h.bot.Send(msg)
}
