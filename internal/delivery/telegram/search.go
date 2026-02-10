package telegram

import (
	"context"
	"fmt"
	"html"
	"strings"
	"unicode/utf16"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/m4xvel/monetych_bot/internal/domain"
	"github.com/m4xvel/monetych_bot/internal/logger"
)

const maxTelegramMessageLen = 3800

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

	summary := h.formatOrderSummary(orderFull)
	if err := h.sendSearchSummary(ctx, chatID, msg.MessageID, summary, mediaCount, orderFull.Order.ID); err != nil {
		return
	}

	if len(orderFull.Messages) > 0 {
		chatLines := make([]string, 0, len(orderFull.Messages))
		for _, message := range orderFull.Messages {
			chatLines = append(chatLines, h.formatChatMessage(message))
		}

		for _, chunk := range h.buildChatChunks(chatLines, maxTelegramMessageLen) {
			response := tgbotapi.NewMessage(chatID, chunk)
			response.ParseMode = tgbotapi.ModeHTML
			if _, err := h.bot.Send(response); err != nil {
				wrapped := wrapTelegramErr("telegram.send_search_result", err)
				logger.Log.Errorw("failed to send order search result chunk",
					"chat_id", chatID,
					"order_id", orderFull.Order.ID,
					"err", wrapped,
				)
				return
			}
		}
	}
}

func (h *Handler) formatOrderSummary(orderFull *domain.OrderFull) string {
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

func (h *Handler) sendSearchSummary(
	ctx context.Context,
	chatID int64,
	replyTo int,
	summary string,
	mediaCount int,
	orderID int,
) error {
	parts := splitByLineLimit(summary, maxTelegramMessageLen)

	for i, part := range parts {
		response := tgbotapi.NewMessage(chatID, part)
		response.ParseMode = tgbotapi.ModeHTML
		if i == 0 {
			response.ReplyToMessageID = replyTo
			if mediaCount > 0 {
				showMediaToken, err := h.callbackTokenService.Create(
					ctx,
					"show_media",
					&SearchPayload{
						ChatID:  chatID,
						OrderID: orderID,
					},
				)
				if err != nil {
					logger.Log.Errorw("failed to create show media callback token",
						"chat_id", chatID,
						"err", err,
					)
				} else {
					showMediaButton := tgbotapi.NewInlineKeyboardButtonData(
						fmt.Sprintf(h.text.SearchShowMediaButtonTemplate, mediaCount),
						"show_media:"+showMediaToken,
					)
					response.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(
						tgbotapi.NewInlineKeyboardRow(showMediaButton),
					)
				}
			}
		}

		if _, err := h.bot.Send(response); err != nil {
			wrapped := wrapTelegramErr("telegram.send_search_result", err)
			logger.Log.Errorw("failed to send order search result",
				"chat_id", chatID,
				"order_id", orderID,
				"err", wrapped,
			)
			return err
		}
	}

	return nil
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

func (h *Handler) buildChatChunks(
	lines []string,
	maxLen int,
) []string {
	if len(lines) == 0 {
		return nil
	}

	header := h.text.SearchChatHeader
	wrapperOpen := "<blockquote expandable>\n"
	wrapperClose := "\n</blockquote>"

	headerLen := tgLen(header)
	wrapperLen := tgLen(wrapperOpen) + tgLen(wrapperClose)

	availableFirst := maxLen - headerLen - wrapperLen
	if availableFirst < 1 {
		availableFirst = maxLen
	}

	availableNext := maxLen - wrapperLen
	if availableNext < 1 {
		availableNext = maxLen
	}

	var chunks []string
	var b strings.Builder
	curLen := 0
	isFirst := true
	available := availableFirst

	flush := func() {
		if curLen == 0 {
			return
		}
		prefix := ""
		if isFirst {
			prefix = header
			isFirst = false
		}
		chunks = append(chunks, prefix+wrapperOpen+b.String()+wrapperClose)
		b.Reset()
		curLen = 0
		available = availableNext
	}

	for _, line := range lines {
		lineLen := tgLen(line)

		if lineLen > available {
			if curLen > 0 {
				flush()
			}

			for _, part := range splitByTgLimit(line, available) {
				if part == "" {
					continue
				}
				if isFirst {
					chunks = append(chunks, header+wrapperOpen+part+wrapperClose)
					isFirst = false
					available = availableNext
					continue
				}
				chunks = append(chunks, wrapperOpen+part+wrapperClose)
			}
			continue
		}

		if curLen+lineLen > available {
			flush()
		}

		b.WriteString(line)
		curLen += lineLen
	}

	if curLen > 0 {
		flush()
	}

	return chunks
}

func splitByLineLimit(text string, limit int) []string {
	if tgLen(text) <= limit {
		return []string{text}
	}

	lines := strings.SplitAfter(text, "\n")
	return splitLinesByLimit(lines, limit)
}

func splitLinesByLimit(lines []string, limit int) []string {
	var chunks []string
	var b strings.Builder
	curLen := 0

	for _, line := range lines {
		if line == "" {
			continue
		}

		lineLen := tgLen(line)
		if lineLen > limit {
			if curLen > 0 {
				chunks = append(chunks, b.String())
				b.Reset()
				curLen = 0
			}
			chunks = append(chunks, splitByTgLimit(line, limit)...)
			continue
		}

		if curLen+lineLen > limit {
			chunks = append(chunks, b.String())
			b.Reset()
			curLen = 0
		}

		b.WriteString(line)
		curLen += lineLen
	}

	if curLen > 0 {
		chunks = append(chunks, b.String())
	}

	return chunks
}

func splitByTgLimit(text string, limit int) []string {
	if limit <= 0 {
		return []string{text}
	}

	var parts []string
	var b strings.Builder
	count := 0

	for _, r := range text {
		rLen := 1
		if r > 0xFFFF {
			rLen = 2
		}

		if count+rLen > limit {
			parts = append(parts, b.String())
			b.Reset()
			count = 0
		}
		b.WriteRune(r)
		count += rLen
	}

	if b.Len() > 0 {
		parts = append(parts, b.String())
	}

	return parts
}

func tgLen(text string) int {
	return len(utf16.Encode([]rune(text)))
}
