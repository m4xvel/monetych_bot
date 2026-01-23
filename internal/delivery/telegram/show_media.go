package telegram

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/m4xvel/monetych_bot/internal/domain"
	"github.com/m4xvel/monetych_bot/internal/logger"
)

func (h *Handler) handleShowMedia(
	ctx context.Context,
	cb *tgbotapi.CallbackQuery,
) {
	chatID := cb.Message.Chat.ID

	edit := tgbotapi.EditMessageReplyMarkupConfig{
		BaseEdit: tgbotapi.BaseEdit{
			ChatID:      chatID,
			MessageID:   cb.Message.MessageID,
			ReplyMarkup: nil,
		},
	}

	h.bot.Send(edit)

	logger.Log.Infow("show media selected",
		"chat_id", chatID,
	)

	parts := strings.Split(cb.Data, ":")
	if len(parts) < 2 {
		logger.Log.Warnw("invalid type callback data",
			"chat_id", chatID,
			"data", cb.Data,
		)
		return
	}

	orderID, err := strconv.Atoi(parts[1])
	if err != nil {
		logger.Log.Warnw("failed to parse order id",
			"chat_id", chatID,
			"value", parts[1],
		)
		return
	}

	of, err := h.orderService.FindByID(context.Background(), orderID)
	if err != nil {
		logger.Log.Warnw("order not found by id",
			"chat_id", chatID,
		)
		return
	}

	logger.Log.Infow("order found by id",
		"chat_id", chatID,
		"order_id", of.Order.ID,
	)

	for _, m := range of.Messages {
		if m.Media == nil {
			continue
		}

		fileID, ok := m.Media["file_id"].(string)
		if !ok {
			continue
		}

		switch m.MessageType {
		case domain.MessagePhoto:
			reply := tgbotapi.NewPhoto(chatID, tgbotapi.FileID(fileID))
			reply.ReplyToMessageID = cb.Message.MessageID
			reply.Caption = formatMediaMessage(m)
			reply.ParseMode = tgbotapi.ModeHTML
			h.bot.Send(reply)
		case domain.MessageVideo:
			reply := tgbotapi.NewVideo(chatID, tgbotapi.FileID(fileID))
			reply.ReplyToMessageID = cb.Message.MessageID
			reply.Caption = formatMediaMessage(m)
			reply.ParseMode = tgbotapi.ModeHTML
			h.bot.Send(reply)
		case domain.MessageDocument:
			reply := tgbotapi.NewDocument(chatID, tgbotapi.FileID(fileID))
			reply.ReplyToMessageID = cb.Message.MessageID
			reply.Caption = formatMediaMessage(m)
			reply.ParseMode = tgbotapi.ModeHTML
			h.bot.Send(reply)
		case domain.MessageVoice:
			reply := tgbotapi.NewVoice(chatID, tgbotapi.FileID(fileID))
			reply.ReplyToMessageID = cb.Message.MessageID
			reply.Caption = formatMediaMessage(m)
			reply.ParseMode = tgbotapi.ModeHTML
			h.bot.Send(reply)
		}
	}

	h.bot.Request(tgbotapi.NewCallback(cb.ID, "Медиа отправлены"))

	logger.Log.Infow("media files sent",
		"chat_id", chatID,
		"order_id", of.Order.ID,
	)
}

func formatMediaMessage(m domain.ChatMessage) string {
	var sender string
	switch m.SenderRole {
	case domain.SenderUser:
		sender = "👤 Пользователь"
	case domain.SenderExpert:
		sender = "🧑‍💼 Эксперт"
	default:
		sender = "⚙️ Система"
	}

	var b strings.Builder

	b.WriteString(fmt.Sprintf(
		"<b>%s</b> <i>%s</i>\n",
		sender,
		m.CreatedAt.Format("02.01 15:04"),
	))

	return b.String()
}
