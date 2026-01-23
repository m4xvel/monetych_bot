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

	token := strings.TrimSpace(msg.CommandArguments())
	if token == "" {
		logger.Log.Warnw("search command called without token",
			"chat_id", chatID,
		)

		if _, err := h.bot.Send(tgbotapi.NewMessage(
			chatID,
			"Укажите токен.\nПример:\n/search ZW6T-HJTK-6WY2",
		)); err != nil {
			logger.Log.Errorw("failed to prompt token for search",
				"chat_id", chatID,
				"err", err,
			)
		}
		return
	}

	result, err := h.orderService.FindByToken(ctx, token)
	if err != nil {
		logger.Log.Warnw("order not found by token",
			"chat_id", chatID,
		)

		if _, err := h.bot.Send(tgbotapi.NewMessage(
			chatID,
			"❌ Ничего не найдено по указанному токену",
		)); err != nil {
			logger.Log.Errorw("failed to send not found message",
				"chat_id", chatID,
				"err", err,
			)
		}
		return
	}

	logger.Log.Infow("order found by token",
		"chat_id", chatID,
		"order_id", result.Order.ID,
	)

	mediaCount := 0
	for _, m := range result.Messages {
		if m.Media != nil {
			if _, ok := m.Media["file_id"].(string); ok {
				mediaCount++
			}
		}
	}

	text := FormatOrderFull(result)

	reply := tgbotapi.NewMessage(msg.Chat.ID, text)
	reply.ParseMode = tgbotapi.ModeHTML
	reply.ReplyToMessageID = msg.MessageID

	if mediaCount > 0 {
		button := tgbotapi.NewInlineKeyboardButtonData(
			fmt.Sprintf("📎 Показать медиа (%d)", mediaCount),
			fmt.Sprintf("show_media:%d", result.Order.ID),
		)

		reply.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(button),
		)
	}

	if _, err := h.bot.Send(reply); err != nil {
		logger.Log.Errorw("failed to send order search result",
			"chat_id", chatID,
			"order_id", result.Order.ID,
			"err", err,
		)
	}
}

func FormatOrderFull(of *domain.OrderFull) string {
	if of == nil {
		return "❌ Ошибка: данные заказа отсутствуют"
	}

	var b strings.Builder

	b.WriteString("🧾 <b>Сделка</b>\n")
	b.WriteString(fmt.Sprintf("Статус: <b>%s</b>\n", formatOrderStatus(of.Order.Status)))
	b.WriteString(fmt.Sprintf(
		"Создан: %s\n",
		of.Order.CreatedAt.Format("02.01.2006 15:04"),
	))
	b.WriteString(fmt.Sprintf(
		"Обновлён: %s\n",
		of.Order.UpdatedAt.Format("02.01.2006 15:04"),
	))
	b.WriteString("\n")

	if of.Game != nil && of.Game.ID != 0 {
		b.WriteString("🎮 <b>Игра</b>\n")
		b.WriteString(fmt.Sprintf(
			"Название: <b>%s</b>\n",
			html.EscapeString(of.Game.Name),
		))
		if of.GameType != nil && of.GameType.ID != 0 {
			b.WriteString(fmt.Sprintf(
				"Тип: <b>%s</b>\n",
				html.EscapeString(of.GameType.Name),
			))
		}
		b.WriteString("\n")
	}

	if of.User != nil && of.User.ID != 0 {
		b.WriteString("👤 <b>Пользователь</b>\n")
		b.WriteString(fmt.Sprintf("Имя: %s\n", html.EscapeString(of.User.Name)))
		b.WriteString(fmt.Sprintf("Chat ID: <code>%d</code>\n", of.User.ChatID))
		if of.User.IsVerified {
			b.WriteString("Верифицирован: ✅\n")
		} else {
			b.WriteString("Верифицирован: ❌\n")
		}
		b.WriteString(fmt.Sprintf("Всего заказов: %d\n", of.User.TotalOrders))
		b.WriteString("\n")
	}

	// --- EXPERT ---
	if of.Expert != nil && of.Expert.ID != 0 {
		b.WriteString("🧑‍💼 <b>Эксперт</b>\n")
		b.WriteString(fmt.Sprintf("Chat ID: <code>%d</code>\n", of.Expert.ChatID))
		if of.Expert.IsActive {
			b.WriteString("Активен: ✅\n")
		} else {
			b.WriteString("Активен: ❌\n")
		}
		b.WriteString("\n")
	}

	// --- USER STATE ---
	if of.UserState != nil && of.UserState.State != "" {
		b.WriteString("📝 <b>Состояние пользователя</b>\n")
		b.WriteString(fmt.Sprintf("State: <b>%s</b>\n", formatStateName(of.UserState.State)))
		b.WriteString(fmt.Sprintf(
			"Обновлено: %s\n",
			of.UserState.UpdatedAt.Format("02.01.2006 15:04"),
		))
	}

	// --- CHAT ---
	if len(of.Messages) > 0 {
		b.WriteString("\n💬 <b>Чат</b>\n")

		var chat strings.Builder

		for _, m := range of.Messages {
			chat.WriteString(formatChatMessage(m))
		}

		if chat.Len() > 0 {
			b.WriteString(collapsibleQuoteHTML(chat.String()))
		}
	}

	return b.String()
}

func formatChatMessage(m domain.ChatMessage) string {
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

	wroteContent := false

	if m.Text != nil && *m.Text != "" {
		b.WriteString(fmt.Sprintf("\t\t\t\t\t\t> %s", *m.Text))
		b.WriteString("\n")
		wroteContent = true
	}

	if m.Media != nil {
		b.WriteString(fmt.Sprintf("\t\t\t\t\t\t> %s",
			formatMedia(m.MessageType, m.Media)))
		wroteContent = true
	}

	if !wroteContent {
		b.WriteString("\t\t\t\t\t\t> 🔡 <b>Другое</b>\n")
	}

	b.WriteString("\n")
	return b.String()
}

func formatOrderStatus(
	status domain.OrderStatus,
) string {

	switch status {

	case domain.OrderNew:
		return "создан"

	case domain.OrderAccepted:
		return "принят"

	case domain.OrderExpertConfirmed:
		return "подтверждён экспертом"

	case domain.OrderCompleted:
		return "подтверждён клиентом"

	case domain.OrderDeclined:
		return "отменён экспертом"

	case domain.OrderCanceled:
		return "отменён клиентом"
	}

	return ""
}

func formatStateName(
	state domain.StateName,
) string {

	switch state {

	case domain.StateIdle:
		return "в ожидании"

	case domain.StateStart:
		return "начало"

	case domain.StateCommunication:
		return "общается с экспертом"

	case domain.StateWritingReview:
		return "пишет отзыв"
	}

	return ""
}

func formatMedia(
	msgType domain.MessageType,
	media map[string]any,
) string {

	switch msgType {
	case domain.MessagePhoto:
		return "🖼 <b>Фото</b>\n"

	case domain.MessageVideo:
		return "🎥 <b>Видео</b>\n"

	case domain.MessageDocument:
		if name, ok := media["file_name"].(string); ok {
			return fmt.Sprintf("📎 <b>Документ</b> : %s\n", name)
		}
		return "📎 <b>Документ</b>\n"

	case domain.MessageVoice:
		return "🎤 <b>Голосовое сообщение</b>\n"
	}

	return ""
}

func collapsibleQuoteHTML(text string) string {
	if text == "" {
		return ""
	}

	return fmt.Sprintf(
		"<blockquote expandable>\n%s\n</blockquote>",
		text,
	)
}
