package telegram

import (
	"context"
	"fmt"
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

	if _, err := h.bot.Send(edit); err != nil {
		wrapped := wrapTelegramErr("telegram.remove_show_media_keyboard", err)
		logger.Log.Errorw("failed to remove show media keyboard",
			"chat_id", chatID,
			"err", wrapped,
		)
	}

	logger.Log.Infow("show media selected",
		"chat_id", chatID,
	)

	parts := strings.Split(cb.Data, ":")
	if len(parts) != 2 {
		logger.Log.Warnw("invalid type callback data",
			"chat_id", chatID,
			"data", cb.Data,
		)
		return
	}

	tokenCallback := parts[1]

	var payload SearchPayload

	if err := h.callbackTokenService.Consume(
		ctx,
		tokenCallback,
		"show_media",
		&payload,
	); err != nil {
		if isInvalidToken(err) {
			logger.Log.Warnw("invalid show media callback token",
				"chat_id", chatID,
				"data", cb.Data,
				"err", err,
			)
			return
		}
		logger.Log.Errorw("failed to consume show media callback token",
			"chat_id", chatID,
			"err", err,
		)
		return
	}

	if payload.ChatID != cb.From.ID {
		return
	}

	orderID := payload.OrderID

	orderFull, err := h.orderService.FindByID(context.Background(), orderID)
	if err != nil {
		logger.Log.Warnw("order not found by id",
			"chat_id", chatID,
		)
		return
	}

	logger.Log.Infow("order found by id",
		"chat_id", chatID,
		"order_id", orderFull.Order.ID,
	)

	for _, chatMessage := range orderFull.Messages {
		if chatMessage.Media == nil {
			continue
		}

		fileID, ok := chatMessage.Media["file_id"].(string)
		if !ok {
			continue
		}

		switch chatMessage.MessageType {
		case domain.MessagePhoto:
			photoReply := tgbotapi.NewPhoto(chatID, tgbotapi.FileID(fileID))
			photoReply.ReplyToMessageID = cb.Message.MessageID
			photoReply.Caption = h.formatMediaCaption(chatMessage)
			photoReply.ParseMode = tgbotapi.ModeHTML
			if _, err := h.bot.Send(photoReply); err != nil {
				wrapped := wrapTelegramErr("telegram.send_media_photo", err)
				logger.Log.Errorw("failed to send photo",
					"chat_id", chatID,
					"order_id", orderID,
					"err", wrapped,
				)
			}
		case domain.MessageVideo:
			videoReply := tgbotapi.NewVideo(chatID, tgbotapi.FileID(fileID))
			videoReply.ReplyToMessageID = cb.Message.MessageID
			videoReply.Caption = h.formatMediaCaption(chatMessage)
			videoReply.ParseMode = tgbotapi.ModeHTML
			if _, err := h.bot.Send(videoReply); err != nil {
				wrapped := wrapTelegramErr("telegram.send_media_video", err)
				logger.Log.Errorw("failed to send video",
					"chat_id", chatID,
					"order_id", orderID,
					"err", wrapped,
				)
			}
		case domain.MessageDocument:
			documentReply := tgbotapi.NewDocument(chatID, tgbotapi.FileID(fileID))
			documentReply.ReplyToMessageID = cb.Message.MessageID
			documentReply.Caption = h.formatMediaCaption(chatMessage)
			documentReply.ParseMode = tgbotapi.ModeHTML
			if _, err := h.bot.Send(documentReply); err != nil {
				wrapped := wrapTelegramErr("telegram.send_media_document", err)
				logger.Log.Errorw("failed to send document",
					"chat_id", chatID,
					"order_id", orderID,
					"err", wrapped,
				)
			}
		case domain.MessageVoice:
			voiceReply := tgbotapi.NewVoice(chatID, tgbotapi.FileID(fileID))
			voiceReply.ReplyToMessageID = cb.Message.MessageID
			voiceReply.Caption = h.formatMediaCaption(chatMessage)
			voiceReply.ParseMode = tgbotapi.ModeHTML
			if _, err := h.bot.Send(voiceReply); err != nil {
				wrapped := wrapTelegramErr("telegram.send_media_voice", err)
				logger.Log.Errorw("failed to send voice",
					"chat_id", chatID,
					"order_id", orderID,
					"err", wrapped,
				)
			}
		}
	}

	h.answerCallback(cb, h.text.MediaSentToast)

	logger.Log.Infow("media files sent",
		"chat_id", chatID,
		"order_id", orderFull.Order.ID,
	)
}

func (h *Handler) formatMediaCaption(chatMessage domain.ChatMessage) string {
	var sender string
	switch chatMessage.SenderRole {
	case domain.SenderUser:
		sender = h.text.SenderUserLabel
	case domain.SenderExpert:
		sender = h.text.SenderExpertLabel
	default:
		sender = h.text.SenderSystemLabel
	}

	var b strings.Builder

	b.WriteString(fmt.Sprintf(
		h.text.ChatMessageHeaderTemplate,
		sender,
		chatMessage.CreatedAt.Format("02.01 15:04"),
	))

	return b.String()
}
