package telegram

import (
	"context"
	"fmt"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/m4xvel/monetych_bot/internal/domain"
)

func (h *Handler) SearchCommand(ctx context.Context, msg *tgbotapi.Message) {
	token := msg.CommandArguments()

	if token == "" {
		h.bot.Send(tgbotapi.NewMessage(
			msg.Chat.ID,
			"Укажите токен.\nПример:\n/search ZW6T-HJTK-6WY2",
		))
		return
	}

	// если нужно — нормализуем
	token = strings.TrimSpace(token)

	result, err := h.orderService.FindByToken(ctx, token)
	if err != nil {
		h.bot.Send(tgbotapi.NewMessage(
			msg.Chat.ID,
			"❌ Ничего не найдено по токену: "+token,
		))
		return
	}

	text := FormatOrderFullMarkdown(result)

	reply := tgbotapi.NewMessage(msg.Chat.ID, text)
	reply.ParseMode = tgbotapi.ModeMarkdown
	reply.ReplyToMessageID = msg.MessageID

	h.bot.Send(reply)
}

func FormatOrderFullMarkdown(of *domain.OrderFull) string {
	if of == nil {
		return "❌ Ошибка: данные заказа отсутствуют"
	}

	var b strings.Builder

	b.WriteString("🧾 *Сделка*\n")
	b.WriteString(fmt.Sprintf("Статус: *%s*\n", of.Order.Status))
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
		b.WriteString("🎮 *Игра*\n")
		b.WriteString(fmt.Sprintf("Название: *%s*\n", escapeMarkdown(of.Game.Name)))
		if of.GameType != nil && of.GameType.ID != 0 {
			b.WriteString(fmt.Sprintf("Тип: *%s*\n", escapeMarkdown(of.GameType.Name)))
		}
		b.WriteString("\n")
	}

	if of.User != nil && of.User.ID != 0 {
		b.WriteString("👤 *Пользователь*\n")
		b.WriteString(fmt.Sprintf("Имя: %s\n", escapeMarkdown(of.User.Name)))
		b.WriteString(fmt.Sprintf("Chat ID: `%d`\n", of.User.ChatID))
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
		b.WriteString("🧑‍💼 *Эксперт*\n")
		b.WriteString(fmt.Sprintf("Chat ID: `%d`\n", of.Expert.ChatID))
		if of.Expert.IsActive {
			b.WriteString("Активен: ✅\n")
		} else {
			b.WriteString("Активен: ❌\n")
		}
		b.WriteString("\n")
	}

	// --- USER STATE ---
	if of.UserState != nil && of.UserState.State != "" {
		b.WriteString("📝 *Состояние пользователя*\n")
		b.WriteString(fmt.Sprintf("State: *%s*\n", of.UserState.State))
		b.WriteString(fmt.Sprintf(
			"Обновлено: %s\n",
			of.UserState.UpdatedAt.Format("02.01.2006 15:04"),
		))
	}

	return b.String()
}

func escapeMarkdown(s string) string {
	replacer := strings.NewReplacer(
		"_", "\\_",
		"*", "\\*",
		"[", "\\[",
		"]", "\\]",
		"`", "\\`",
	)
	return replacer.Replace(s)
}
