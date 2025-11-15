package utils

import "fmt"

type Dynamic struct{}

type Messages struct {
	ChooseGame                string
	ChooseType                string
	AlreadyActiveOrder        string
	YouNeedToVerify           string
	ContactAppraiserText      string
	ContactText               string
	VerifyButtonText          string
	SuccessfulVerify          string
	FailedVerify              string
	WaitingAssessor           string
	AssessorAcceptedYourOrder string
	AcceptText                string
	DeclineText               string
	ApplicationManagementText string
	ConfirmYourOrder          string
	YouHaveCancelledOrder     string
	YouOrderCancelled         string
	OrderConfirmed            string
	YouConfirmedPayment       string
}

func NewMessages() *Messages {
	return &Messages{
		ChooseGame:                "Выберите игру 🎮",
		ChooseType:                "Выберите тип 📦",
		AlreadyActiveOrder:        "У вас уже есть активная заявка!",
		YouNeedToVerify:           "Для безопасной сделки необходимо подтвердить вашу личность. Это просто и не займет много времени:",
		ContactAppraiserText:      "Свяжитесь с оценщиком 📩",
		ContactText:               "Связаться 💬",
		VerifyButtonText:          "Пройти верификацию",
		SuccessfulVerify:          "✅ Ваша личность подтверждена!",
		FailedVerify:              "❌ Вы не прошли верификацию, попробуйте снова!",
		WaitingAssessor:           "⏳ Оценщик уже спешит к Вам",
		AssessorAcceptedYourOrder: "✅ Оценщик принял Вашу заявку, продолжайте общаться в этом чате!",
		AcceptText:                "✅ Подтвердить",
		DeclineText:               "❌ Отклонить",
		ApplicationManagementText: "⚙ Управление заявкой",
		ConfirmYourOrder:          "Заказ оплачен, проверьте счет и подтвердите заказ!",
		YouHaveCancelledOrder:     "Вы отменили заказ!",
		YouOrderCancelled:         "Ваш заказ отменен!",
		OrderConfirmed:            "Заказ подтвержден!",
		YouConfirmedPayment:       "✅ Вы подтвердили оплату",
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

func (d *Dynamic) TitleOrderTopic(orderID int, itemGame, itemType string) string {
	return fmt.Sprintf("💼 Сделка #%d - (%s, %s)", orderID, itemGame, itemType)
}
