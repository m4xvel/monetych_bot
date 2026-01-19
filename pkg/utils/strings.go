package utils

import (
	"fmt"
	"strings"
)

type Dynamic struct{}

type Messages struct {
	HelloText             string
	ChooseGame            string
	ChooseType            string
	AlreadyActiveOrder    string
	YouNeedToVerify       string
	ContactAppraiserText  string
	ContactText           string
	VerifyButtonText      string
	SuccessfulVerify      string
	FailedVerify          string
	WaitingAssessor       string
	AcceptText            string
	DeclineText           string
	ConfirmYourOrder      string
	YouHaveCancelledOrder string
	YouConfirmedOrder     string
	YouOrderCancelled     string
	OrderConfirmed        string
	YouConfirmedPayment   string
	SupportText           string
	StartMenuText         string
	CatalogMenuText       string
	ConfirmDeclineText    string
	ConfirmConfirmedText  string
	SupportMenuText       string
	ReviewsMenuText       string
	ThanksForReviewText   string
	ChatClosedText        string
	WriteReviewText       string
}

func NewMessages() *Messages {
	return &Messages{
		HelloText:             "Привет! 👋 Добро пожаловать в наш сервис.\n\nЗдесь ты можешь быстро и безопасно продать игровой аккаунт, предмет, скин или любую ценность из игры.\nМы делаем всё максимально просто: выбираешь игру, проходишь короткие шаги — и наш эксперт поможет со всем дальше.\nЕсли что-то будет непонятно, мы всегда рядом 😊\n\nГотов начать? 🎮✨",
		ChooseGame:            "Выбери игру 🎮",
		ChooseType:            "Выбери, что хочешь продать ✨",
		AlreadyActiveOrder:    "У тебя уже есть активная заявка 🙂",
		YouNeedToVerify:       "Чтобы сделать сделку безопасной и быстрой, нужно пройти короткую верификацию. Это займёт меньше минуты:",
		ContactAppraiserText:  "Наш эксперт уже готов помочь с оценкой!",
		ContactText:           "Связаться с экспертом 💬",
		VerifyButtonText:      "Пройти верификацию ✔️",
		SuccessfulVerify:      "Отлично! Ты прошёл верификацию 🎉",
		FailedVerify:          "Не получилось пройти верификацию 😕\nПопробуй ещё раз - это важно для безопасности сделки",
		WaitingAssessor:       "Ждём эксперта ⏳\nОбычно это занимает 1–3 минуты.",
		AcceptText:            "Подтвердить ✔️",
		DeclineText:           "Отменить ❌",
		ConfirmYourOrder:      "Эксперт отправил оплату 💸\nПроверь счёт - если всё верно, подтверди получение.",
		YouHaveCancelledOrder: "Ты отменил заявку 🚫",
		YouConfirmedOrder:     "Ты подтвердил выполнение заказа!",
		YouOrderCancelled:     "Эксперт отменил заявку 😕",
		OrderConfirmed:        "Клиент подтвердил сделку ✔️",
		YouConfirmedPayment:   "Ты подтвердил получение выплаты 🎉\nСпасибо, что выбрал наш сервис! 😊",
		SupportText:           "Служба поддержки: @support",
		StartMenuText:         "♻ Обновить меню",
		CatalogMenuText:       "🎮 Открыть каталог",
		ConfirmDeclineText:    "⚠️ Подтвердите действие\n\nВы действительно хотите отменить заказ?",
		ConfirmConfirmedText:  "⚠️ Подтвердите действие\n\nВы действительно хотите подтвердить выполнение заказа?",
		SupportMenuText:       "👨‍💻 Поддержка",
		ReviewsMenuText:       "⭐️ Отзывы клиентов",
		ThanksForReviewText:   "Спасибо за отзыв!",
		ChatClosedText:        "Чат закрыт! Оцените наш сервис от 1 до 5 ⭐",
		WriteReviewText:       "Теперь напишите ваш отзыв ✍️",
	}
}

func NewDynamic() *Dynamic {
	return &Dynamic{}
}

func (d *Dynamic) YouHaveChosenGame(gameName string) string {
	return fmt.Sprintf("🎮 Вы выбрали: %s", gameName)
}

func (d *Dynamic) YouHaveChosenType(itemType string) string {
	return fmt.Sprintf("📦 Вы выбрали: %s", itemType)
}

func (d *Dynamic) NewOrder(orderID int, nameGame, nameType string) string {
	return fmt.Sprintf("Новая заявка #%d: %s, %s", orderID, nameGame, nameType)
}

func (d *Dynamic) AssessorAcceptedOrder(orderID int, itemGame, itemType string) string {
	return fmt.Sprintf(
		"Вы приняли заявку #%d ✅\n(%s, %s)",
		orderID, itemGame, itemType,
	)
}

func (d *Dynamic) AssessorAcceptedYourOrder(token string) string {
	return fmt.Sprintf(
		"Токен для обращения в поддержку\n\n`%s`\n\nЭксперт принял твою заявку! 🎉\nДальнейшее общение будет прямо здесь — удобно и быстро 😌",
		escapeMarkdown(token),
	)
}

func (d *Dynamic) TitleOrderTopic(orderID int, itemGame, itemType string) string {
	return fmt.Sprintf("💼 Сделка #%d - (%s, %s)", orderID, itemGame, itemType)
}

func (d *Dynamic) ApplicationManagementText(gameName,
	gameTypeName string) string {
	return fmt.Sprintf(
		"Панель управления заявкой ⚙️\n\nИгра: %s\nТип: %s",
		gameName,
		gameTypeName,
	)
}

func escapeMarkdown(s string) string {
	replacer := strings.NewReplacer(
		"_", "\\_",
		"*", "\\*",
		"[", "\\[",
		"]", "\\]",
		"(", "\\(",
		")", "\\)",
		"~", "\\~",
		"`", "\\`",
		">", "\\>",
		"#", "\\#",
		"+", "\\+",
		"-", "\\-",
		"=", "\\=",
		"|", "\\|",
		"{", "\\{",
		"}", "\\}",
		".", "\\.",
		"!", "\\!",
	)
	return replacer.Replace(s)
}
