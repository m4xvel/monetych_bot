package telegram

import (
	"context"
	"encoding/json"
	"fmt"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/m4xvel/monetych_bot/internal/domain"
	"github.com/m4xvel/monetych_bot/internal/logger"
)

func (h *Handler) handleMessage(
	ctx context.Context,
	msg *tgbotapi.Message,
) {
	if msg.From != nil && msg.From.IsBot {
		return
	}

	if msg.MessageThreadID != 0 {
		h.handleExpertMessage(ctx, msg)
		return
	}

	h.handleUserMessage(ctx, msg)
}

func (h *Handler) handleUserMessage(
	ctx context.Context,
	msg *tgbotapi.Message,
) {
	chatID := msg.Chat.ID

	state, err := h.stateService.GetStateByChatID(ctx, chatID)
	if err != nil {
		logger.Log.Errorw("failed to get user state",
			"chat_id", chatID,
			"err", err,
		)
		return
	}
	if state == nil {
		logger.Log.Warnw("user state not found",
			"chat_id", chatID,
		)
		return
	}

	switch state.State {

	case domain.StateCommunication:
		logger.Log.Infow("user message forwarded to expert",
			"chat_id", chatID,
			"order_id", state.OrderID,
		)

		text := extractText(msg)

		media, msgType := extractMedia(msg)

		if err := h.orderChatMessageService.SaveUserMessage(
			ctx,
			*state.OrderID,
			*state.UserID,
			msg.Chat.ID,
			msg.MessageID,
			msgType,
			text,
			media,
		); err != nil {
			logger.Log.Errorw("failed to save user message",
				"err", err,
			)
			return
		}

		logger.Log.Infow("user message data save in database",
			"order_id", state.OrderID,
		)

		h.forwardToExpert(state, msg)

	case domain.StateStart:
		logger.Log.Infow("user message in start state redirected to catalog",
			"chat_id", chatID,
		)
		h.handlerCatalogCommand(ctx, msg)

	case domain.StateWritingReview:
		var text string
		switch {
		case msg.Text != "":
			text = msg.Text
		case msg.Caption != "":
			text = msg.Caption
		default:
			return
		}

		if err := h.reviewService.AddText(ctx, *state.ReviewID, text); err != nil {
			logger.Log.Errorw("failed to add review text",
				"chat_id", chatID,
				"review_id", *state.ReviewID,
				"err", err,
			)
			return
		}

		if err := h.reviewService.Publish(ctx, *state.ReviewID); err != nil {
			logger.Log.Errorw("failed to publish review",
				"chat_id", chatID,
				"review_id", *state.ReviewID,
				"err", err,
			)
			return
		}

		logger.Log.Infow("review published",
			"chat_id", chatID,
			"review_id", *state.ReviewID,
		)

		if err := h.stateService.SetStateIdle(ctx, chatID); err != nil {
			logger.Log.Errorw("failed to set idle state after review",
				"chat_id", chatID,
				"err", err,
			)
		}

		h.bot.Send(tgbotapi.NewMessage(chatID, h.text.ThanksForReviewText))
	}
}

func (h *Handler) forwardToExpert(
	state *domain.UserState,
	msg *tgbotapi.Message,
) {
	if state.ExpertTopicID == nil || state.OrderThreadID == nil {
		return
	}

	params := tgbotapi.Params{
		"chat_id":           int64PtrToStr(state.ExpertTopicID),
		"from_chat_id":      fmt.Sprint(msg.Chat.ID),
		"message_id":        fmt.Sprint(msg.MessageID),
		"message_thread_id": int64PtrToStr(state.OrderThreadID),
	}

	h.bot.MakeRequest("copyMessage", params)
}

func (h *Handler) handleExpertMessage(
	ctx context.Context,
	msg *tgbotapi.Message,
) {
	if isSystemMessage(msg) {
		return
	}

	state, err := h.stateService.GetStateByThreadID(ctx, msg.MessageThreadID)
	if err != nil {
		logger.Log.Errorw("failed to get state by thread id",
			"thread_id", msg.MessageThreadID,
			"err", err,
		)
		return
	}
	if state == nil {
		logger.Log.Warnw("state not found for expert message",
			"thread_id", msg.MessageThreadID,
		)
		return
	}

	if !canExpertWrite(*state.OrderStatus) {
		logger.Log.Warnw("expert message blocked by order status",
			"order_id", state.OrderID,
			"status", state.OrderStatus,
		)
		return
	}

	text := extractText(msg)

	media, msgType := extractMedia(msg)

	if err := h.orderChatMessageService.SaveExpertMessage(
		ctx,
		*state.OrderID,
		*state.ExpertID,
		msg.Chat.ID,
		msg.MessageID,
		msgType,
		text,
		media,
	); err != nil {
		logger.Log.Errorw("failed to save expert message",
			"err", err,
		)
		return
	}

	logger.Log.Infow("expert message data save in database",
		"order_id", state.OrderID,
	)

	params := tgbotapi.Params{
		"chat_id":      int64PtrToStr(state.UserChatID),
		"from_chat_id": fmt.Sprint(msg.Chat.ID),
		"message_id":   fmt.Sprint(msg.MessageID),
	}

	if _, err := h.bot.MakeRequest("copyMessage", params); err != nil {
		logger.Log.Errorw("failed to forward expert message to user",
			"order_id", state.OrderID,
			"user_chat_id", state.UserChatID,
			"err", err,
		)
		return
	}

	logger.Log.Infow("expert message forwarded to user",
		"order_id", state.OrderID,
	)
}

func extractText(msg *tgbotapi.Message) *string {
	switch {
	case msg.Text != "":
		return &msg.Text
	case msg.Caption != "":
		return &msg.Caption
	default:
		return nil
	}
}

func extractMedia(
	msg *tgbotapi.Message,
) (map[string]any, domain.MessageType) {
	switch {
	case len(msg.Photo) > 0:
		p := msg.Photo[len(msg.Photo)-1]
		return map[string]any{
			"file_id":        p.FileID,
			"file_unique_id": p.FileUniqueID,
			"width":          p.Width,
			"height":         p.Height,
		}, domain.MessagePhoto

	case msg.Document != nil:
		return map[string]any{
			"file_id":        msg.Document.FileID,
			"file_unique_id": msg.Document.FileUniqueID,
			"file_name":      msg.Document.FileName,
			"mime_type":      msg.Document.MimeType,
			"file_size":      msg.Document.FileSize,
		}, domain.MessageDocument

	case msg.Video != nil:
		return map[string]any{
			"file_id":        msg.Video.FileID,
			"file_unique_id": msg.Video.FileUniqueID,
			"duration":       msg.Video.Duration,
			"mime_type":      msg.Video.MimeType,
		}, domain.MessageVideo

	case msg.Voice != nil:
		return map[string]any{
			"file_id":        msg.Voice.FileID,
			"file_unique_id": msg.Voice.FileUniqueID,
			"duration":       msg.Voice.Duration,
		}, domain.MessageVoice
	}

	if msg.Text != "" || msg.Caption != "" {
		return nil, domain.MessageText
	}

	return nil, domain.MessageOther
}

func structToMap(v any) (map[string]any, error) {
	b, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}

	var m map[string]any
	err = json.Unmarshal(b, &m)
	return m, err
}

func int64PtrToStr(v *int64) string {
	if v == nil {
		return ""
	}
	return fmt.Sprint(*v)
}

func isSystemMessage(msg *tgbotapi.Message) bool {
	return msg.NewChatMembers != nil ||
		msg.LeftChatMember != nil ||
		msg.PinnedMessage != nil ||
		msg.NewChatTitle != "" ||
		msg.DeleteChatPhoto ||
		msg.GroupChatCreated ||
		msg.SuperGroupChatCreated ||
		msg.ChannelChatCreated
}

func canExpertWrite(status domain.OrderStatus) bool {
	switch status {
	case domain.OrderAccepted, domain.OrderExpertConfirmed:
		return true
	default:
		return false
	}
}
