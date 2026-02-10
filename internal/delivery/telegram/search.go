package telegram

import (
	"context"
	"fmt"
	"html"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/m4xvel/monetych_bot/internal/domain"
	"github.com/m4xvel/monetych_bot/internal/logger"
)

func (h *Handler) SearchCommand(ctx context.Context, msg *tgbotapi.Message) {
	chatID := msg.Chat.ID

	logger.Log.Infow("search command initiated",
		"chat_id", chatID,
	)

	searchToken := strings.TrimSpace(msg.CommandArguments())
	if searchToken == "" {
		logger.Log.Warnw("search command called without token",
			"chat_id", chatID,
		)

		if _, err := h.bot.Send(tgbotapi.NewMessage(
			chatID,
			h.text.SearchTokenPromptText,
		)); err != nil {
			wrapped := wrapTelegramErr("telegram.send_search_prompt", err)
			logger.Log.Errorw("failed to prompt token for search",
				"chat_id", chatID,
				"err", wrapped,
			)
		}
		return
	}

	orderFull, err := h.orderService.FindByToken(ctx, searchToken)
	if err != nil {
		logger.Log.Warnw("order not found by token",
			"chat_id", chatID,
		)

		if _, err := h.bot.Send(tgbotapi.NewMessage(
			chatID,
			h.text.SearchNotFoundText,
		)); err != nil {
			wrapped := wrapTelegramErr("telegram.send_search_not_found", err)
			logger.Log.Errorw("failed to send not found message",
				"chat_id", chatID,
				"err", wrapped,
			)
		}
		return
	}

	logger.Log.Infow("order found by token",
		"chat_id", chatID,
		"order_id", orderFull.Order.ID,
	)

	mediaCount := 0
	for _, message := range orderFull.Messages {
		if message.Media != nil {
			if _, ok := message.Media["file_id"].(string); ok {
				mediaCount++
			}
		}
	}

	formattedText := h.formatOrderFull(orderFull)

	response := tgbotapi.NewMessage(msg.Chat.ID, formattedText)
	response.ParseMode = tgbotapi.ModeHTML
	response.ReplyToMessageID = msg.MessageID

	if mediaCount > 0 {

		showMediaToken, err := h.callbackTokenService.Create(
			ctx,
			"show_media",
			&SearchPayload{
				ChatID:  chatID,
				OrderID: orderFull.Order.ID,
			},
		)
		if err != nil {
			logger.Log.Errorw("failed to create show media callback token",
				"chat_id", chatID,
				"err", err,
			)
		}

		showMediaButton := tgbotapi.NewInlineKeyboardButtonData(
			fmt.Sprintf(h.text.SearchShowMediaButtonTemplate, mediaCount),
			"show_media:"+showMediaToken,
		)

		response.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(showMediaButton),
		)
	}

	if _, err := h.bot.Send(response); err != nil {
		wrapped := wrapTelegramErr("telegram.send_search_result", err)
		logger.Log.Errorw("failed to send order search result",
			"chat_id", chatID,
			"order_id", orderFull.Order.ID,
			"err", wrapped,
		)
	}
}

func (h *Handler) formatOrderFull(orderFull *domain.OrderFull) string {
	if orderFull == nil {
		return h.text.SearchMissingOrderText
	}

	var builder strings.Builder

	builder.WriteString(h.text.SearchDealHeader)
	builder.WriteString(fmt.Sprintf(
		h.text.SearchStatusLineTemplate,
		h.formatOrderStatus(orderFull.Order.Status),
	))
	builder.WriteString(fmt.Sprintf(
		h.text.SearchCreatedLineTemplate,
		orderFull.Order.CreatedAt.Format("02.01.2006 15:04"),
	))
	builder.WriteString(fmt.Sprintf(
		h.text.SearchUpdatedLineTemplate,
		orderFull.Order.UpdatedAt.Format("02.01.2006 15:04"),
	))
	builder.WriteString("\n")

	if orderFull.Game != nil && orderFull.Game.ID != 0 {
		builder.WriteString(h.text.SearchGameHeader)
		builder.WriteString(fmt.Sprintf(
			h.text.SearchGameNameLineTemplate,
			html.EscapeString(orderFull.Game.Name),
		))
		if orderFull.GameType != nil && orderFull.GameType.ID != 0 {
			builder.WriteString(fmt.Sprintf(
				h.text.SearchGameTypeLineTemplate,
				html.EscapeString(orderFull.GameType.Name),
			))
		}
		builder.WriteString("\n")
	}

	if orderFull.User != nil && orderFull.User.ID != 0 {
		builder.WriteString(h.text.SearchUserHeader)
		builder.WriteString(fmt.Sprintf(
			h.text.SearchUserNameLineTemplate,
			html.EscapeString(orderFull.User.Name),
		))
		builder.WriteString(fmt.Sprintf(
			h.text.SearchUserChatIDLineTemplate,
			orderFull.User.ChatID,
		))
		if orderFull.User.IsVerified {
			builder.WriteString(h.text.SearchUserVerifiedYes)
		} else {
			builder.WriteString(h.text.SearchUserVerifiedNo)
		}
		builder.WriteString(fmt.Sprintf(
			h.text.SearchUserTotalOrdersLineTemplate,
			orderFull.User.TotalOrders,
		))
		builder.WriteString("\n")
	}

	// --- EXPERT ---
	if orderFull.Expert != nil && orderFull.Expert.ID != 0 {
		builder.WriteString(h.text.SearchExpertHeader)
		builder.WriteString(fmt.Sprintf(
			h.text.SearchExpertChatIDLineTemplate,
			orderFull.Expert.TopicID,
		))
		if orderFull.Expert.IsActive {
			builder.WriteString(h.text.SearchExpertActiveYes)
		} else {
			builder.WriteString(h.text.SearchExpertActiveNo)
		}
		builder.WriteString("\n")
	}

	// --- USER STATE ---
	if orderFull.UserState != nil && orderFull.UserState.State != "" {
		builder.WriteString(h.text.SearchUserStateHeader)
		builder.WriteString(fmt.Sprintf(
			h.text.SearchUserStateLineTemplate,
			h.formatStateName(orderFull.UserState.State),
		))
		builder.WriteString(fmt.Sprintf(
			h.text.SearchUserStateUpdatedLineTemplate,
			orderFull.UserState.UpdatedAt.Format("02.01.2006 15:04"),
		))
	}

	// --- CHAT ---
	if len(orderFull.Messages) > 0 {
		builder.WriteString(h.text.SearchChatHeader)

		var chatBuilder strings.Builder

		for _, message := range orderFull.Messages {
			chatBuilder.WriteString(h.formatChatMessage(message))
		}

		if chatBuilder.Len() > 0 {
			builder.WriteString(h.collapsibleQuoteHTML(chatBuilder.String()))
		}
	}

	return builder.String()
}

func (h *Handler) formatChatMessage(chatMessage domain.ChatMessage) string {
	var sender string
	switch chatMessage.SenderRole {
	case domain.SenderUser:
		sender = h.text.SenderUserLabel
	case domain.SenderExpert:
		sender = h.text.SenderExpertLabel
	default:
		sender = h.text.SenderSystemLabel
	}

	var builder strings.Builder

	builder.WriteString(fmt.Sprintf(
		h.text.ChatMessageHeaderTemplate,
		sender,
		chatMessage.CreatedAt.Format("02.01 15:04"),
	))

	wroteContent := false

	if chatMessage.Text != nil && *chatMessage.Text != "" {
		builder.WriteString(fmt.Sprintf(h.text.ChatTextLineTemplate, *chatMessage.Text))
		builder.WriteString("\n")
		wroteContent = true
	}

	if chatMessage.Media != nil {
		builder.WriteString(fmt.Sprintf(
			h.text.ChatTextLineTemplate,
			h.formatMedia(chatMessage.MessageType, chatMessage.Media),
		))
		wroteContent = true
	}

	if !wroteContent {
		builder.WriteString(h.text.ChatOtherLine)
	}

	builder.WriteString("\n")
	return builder.String()
}

func (h *Handler) formatOrderStatus(
	status domain.OrderStatus,
) string {

	switch status {

	case domain.OrderNew:
		return h.text.OrderStatusCreatedText

	case domain.OrderAccepted:
		return h.text.OrderStatusAcceptedText

	case domain.OrderExpertConfirmed:
		return h.text.OrderStatusExpertConfirmedText

	case domain.OrderCompleted:
		return h.text.OrderStatusCompletedText

	case domain.OrderDeclined:
		return h.text.OrderStatusDeclinedByExpertText

	case domain.OrderCanceled:
		return h.text.OrderStatusCanceledByUserText
	}

	return ""
}

func (h *Handler) formatStateName(
	state domain.StateName,
) string {

	switch state {

	case domain.StateIdle:
		return h.text.UserStateIdleText

	case domain.StateStart:
		return h.text.UserStateStartText

	case domain.StateCommunication:
		return h.text.UserStateCommunicationText

	case domain.StateWritingReview:
		return h.text.UserStateWritingReviewText
	}

	return ""
}

func (h *Handler) formatMedia(
	msgType domain.MessageType,
	media map[string]any,
) string {

	switch msgType {
	case domain.MessagePhoto:
		return h.text.MediaPhotoLabel

	case domain.MessageVideo:
		return h.text.MediaVideoLabel

	case domain.MessageVideoNote:
		return h.text.MediaVideoNoteLabel

	case domain.MessageDocument:
		if name, ok := media["file_name"].(string); ok {
			return fmt.Sprintf(h.text.MediaDocumentWithNameTemplate, name)
		}
		return h.text.MediaDocumentLabel

	case domain.MessageVoice:
		return h.text.MediaVoiceLabel
	}

	return ""
}

func (h *Handler) collapsibleQuoteHTML(text string) string {
	if text == "" {
		return ""
	}

	return fmt.Sprintf(
		h.text.ChatQuoteBlockTemplate,
		text,
	)
}
