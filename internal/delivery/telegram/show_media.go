package telegram

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/m4xvel/monetych_bot/internal/domain"
	"github.com/m4xvel/monetych_bot/internal/logger"
)

const (
	mediaBatchSize  = 10
	mediaBatchDelay = 1 * time.Second
	maxSendAttempts = 3
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

	sentCount := 0
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
			if err := h.sendWithRetry("telegram.send_media_photo", chatID, orderID, photoReply); err != nil {
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
			if err := h.sendWithRetry("telegram.send_media_video", chatID, orderID, videoReply); err != nil {
				wrapped := wrapTelegramErr("telegram.send_media_video", err)
				logger.Log.Errorw("failed to send video",
					"chat_id", chatID,
					"order_id", orderID,
					"err", wrapped,
				)
			}
		case domain.MessageVideoNote:
			length, _ := mediaInt(chatMessage.Media, "length")
			videoNoteReply := tgbotapi.NewVideoNote(
				chatID,
				length,
				tgbotapi.FileID(fileID),
			)
			videoNoteReply.ReplyToMessageID = cb.Message.MessageID
			if err := h.sendWithRetry("telegram.send_media_video_note", chatID, orderID, videoNoteReply); err != nil {
				wrapped := wrapTelegramErr("telegram.send_media_video_note", err)
				logger.Log.Errorw("failed to send video note",
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
			if err := h.sendWithRetry("telegram.send_media_document", chatID, orderID, documentReply); err != nil {
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
			if err := h.sendWithRetry("telegram.send_media_voice", chatID, orderID, voiceReply); err != nil {
				wrapped := wrapTelegramErr("telegram.send_media_voice", err)
				logger.Log.Errorw("failed to send voice",
					"chat_id", chatID,
					"order_id", orderID,
					"err", wrapped,
				)
			}
		}

		sentCount++
		if sentCount%mediaBatchSize == 0 {
			time.Sleep(mediaBatchDelay)
		}
	}

	h.answerCallback(cb, h.text.MediaSentToast)

	logger.Log.Infow("media files sent",
		"chat_id", chatID,
		"order_id", orderFull.Order.ID,
	)
}

func mediaInt(media map[string]any, key string) (int, bool) {
	v, ok := media[key]
	if !ok || v == nil {
		return 0, false
	}

	switch value := v.(type) {
	case int:
		return value, true
	case int64:
		return int(value), true
	case float64:
		return int(value), true
	default:
		return 0, false
	}
}

func (h *Handler) sendWithRetry(
	op string,
	chatID int64,
	orderID int,
	msg tgbotapi.Chattable,
) error {
	var lastErr error

	for attempt := 1; attempt <= maxSendAttempts; attempt++ {
		if _, err := h.bot.Send(msg); err == nil {
			return nil
		} else {
			lastErr = err
			retryAfter, ok := retryAfterSeconds(err)
			if !ok || attempt == maxSendAttempts {
				return lastErr
			}
			logger.Log.Warnw("rate limited, retrying send",
				"op", op,
				"chat_id", chatID,
				"order_id", orderID,
				"retry_after", retryAfter,
				"attempt", attempt,
			)
			time.Sleep(time.Duration(retryAfter) * time.Second)
		}
	}

	return lastErr
}

func retryAfterSeconds(err error) (int, bool) {
	var tgErr tgbotapi.Error
	if errors.As(err, &tgErr) && tgErr.Code == 429 && tgErr.RetryAfter > 0 {
		return tgErr.RetryAfter, true
	}
	return 0, false
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
