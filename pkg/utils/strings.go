package utils

import (
	"fmt"
)

type Dynamic struct {
	privacyPolicyURL string
	publicOfferURL   string
}

type Messages struct {
	ChooseGame             string
	ChooseType             string
	AlreadyActiveOrder     string
	YouNeedToVerify        string
	ContactAppraiserText   string
	ContactText            string
	VerifyButtonText       string
	SendToVerificationText string
	SuccessfulVerify       string
	FailedVerify           string
	WaitingAssessor        string
	AcceptText             string
	DeclineText            string
	ConfirmYourOrder       string
	YouHaveCancelledOrder  string
	YouConfirmedOrder      string
	YouOrderCancelled      string
	OrderConfirmed         string
	YouConfirmedPayment    string
	StartMenuText          string
	CatalogMenuText        string
	ConfirmDeclineText     string
	ConfirmConfirmedText   string
	SupportMenuText        string
	ReviewsMenuText        string
	ThanksForReviewText    string
	ChatClosedText         string
	WriteReviewText        string

	AgreeButtonText                    string
	BackButtonText                     string
	AcceptOrderButtonText              string
	SupportContactTemplate             string
	CommunicationBlockedCommandText    string
	CommunicationBlockedCallbackText   string
	NeedAcceptRulesText                string
	VerificationRequestSentToast       string
	VerificationRequestReceivedToast   string
	MediaSentToast                     string
	SearchTokenPromptText              string
	SearchNotFoundText                 string
	SearchShowMediaButtonTemplate      string
	SearchMissingOrderText             string
	SearchDealHeader                   string
	SearchStatusLineTemplate           string
	SearchCreatedLineTemplate          string
	SearchUpdatedLineTemplate          string
	SearchGameHeader                   string
	SearchGameNameLineTemplate         string
	SearchGameTypeLineTemplate         string
	SearchUserHeader                   string
	SearchUserNameLineTemplate         string
	SearchUserChatIDLineTemplate       string
	SearchUserVerifiedYes              string
	SearchUserVerifiedNo               string
	SearchUserTotalOrdersLineTemplate  string
	SearchExpertHeader                 string
	SearchExpertChatIDLineTemplate     string
	SearchExpertActiveYes              string
	SearchExpertActiveNo               string
	SearchUserStateHeader              string
	SearchUserStateLineTemplate        string
	SearchUserStateUpdatedLineTemplate string
	SearchChatHeader                   string
	SenderUserLabel                    string
	SenderExpertLabel                  string
	SenderSystemLabel                  string
	ChatMessageHeaderTemplate          string
	ChatTextLineTemplate               string
	ChatOtherLine                      string
	ChatQuoteBlockTemplate             string
	OrderStatusCreatedText             string
	OrderStatusAcceptedText            string
	OrderStatusExpertConfirmedText     string
	OrderStatusCompletedText           string
	OrderStatusDeclinedByExpertText    string
	OrderStatusCanceledByUserText      string
	UserStateIdleText                  string
	UserStateStartText                 string
	UserStateCommunicationText         string
	UserStateWritingReviewText         string
	MediaPhotoLabel                    string
	MediaVideoLabel                    string
	MediaVideoNoteLabel                string
	MediaDocumentWithNameTemplate      string
	MediaDocumentLabel                 string
	MediaVoiceLabel                    string
}

func NewMessages(privacyPolicyURL, publicOfferURL string) *Messages {
	return &Messages{
		ChooseGame:             "*[Шаг 1/3]* Выбери игру:",
		ChooseType:             "*[Шаг 2/3]* Выбери, что хочешь продать:",
		AlreadyActiveOrder:     "У тебя уже есть активная заявка 🙂",
		YouNeedToVerify:        "Для безопасности сделки нужна быстрая верификация - это займёт меньше минуты ⚡",
		ContactAppraiserText:   "*[Шаг 3/3]* Свяжись с экспертом:",
		ContactText:            "Связаться с экспертом 💬",
		VerifyButtonText:       "Пройти верификацию ✔️",
		SendToVerificationText: "Отправить на верификацию",
		SuccessfulVerify:       "Отлично! Верификация пройдена 🎉\n\nТеперь можно продолжить.",
		FailedVerify:           "Что-то пошло не так... 😕\n\nПопробуй ещё раз!",
		WaitingAssessor:        "Ждём эксперта ⏳\n\nОбычно это занимает 1–3 минуты.",
		AcceptText:             "Подтвердить ✔️",
		DeclineText:            "Отменить ❌",
		ConfirmYourOrder:       "Деньги отправлены! 💸\nПроверь счёт - если всё верно, подтверди получение.",
		YouHaveCancelledOrder:  "Ты отменил заявку 🚫",
		YouConfirmedOrder:      "Ты подтвердил выполнение заказа!",
		YouOrderCancelled:      "Эксперт отменил заявку 😕\n\nЕсли есть вопросы - напиши в поддержку.",
		OrderConfirmed:         "Клиент подтвердил получение ✅",
		YouConfirmedPayment:    "Всё готово! 🎉\n\nДеньги у тебя, мы рады помочь. Возвращайся, если захочешь продать ещё что-нибудь 😊",
		StartMenuText:          "♻ Обновить меню",
		CatalogMenuText:        "🎮 Открыть каталог",
		ConfirmDeclineText:     "⚠️ Подтвердите действие\n\nВы действительно хотите отменить заказ?",
		ConfirmConfirmedText:   "⚠️ Подтвердите действие\n\nВы действительно хотите подтвердить выполнение заказа?",
		SupportMenuText:        "👨‍💻 Поддержка",
		ReviewsMenuText:        "⭐️ Отзывы клиентов",
		ThanksForReviewText:    "Спасибо за отзыв! Это очень важно для нас 🙏",
		ChatClosedText:         "Чат завершён!\n\nОцени наш сервис от 1 до 5 ⭐",
		WriteReviewText:        "Теперь напишите ваш отзыв ✍️",

		AgreeButtonText:                  "Соглашаюсь",
		BackButtonText:                   "⬅️ Вернуться назад",
		AcceptOrderButtonText:            "Принять",
		SupportContactTemplate:           "Поддержка: %s",
		CommunicationBlockedCommandText:  "Вы уже общаетесь с экспертом.\nИспользуйте чат или дождитесь завершения заказа.",
		CommunicationBlockedCallbackText: "Эта кнопка недоступна во время общения с экспертом",
		NeedAcceptRulesText: fmt.Sprintf(
			"Чтобы продолжить работу с ботом, необходимо принять [Публичную оферту](%s) и [Политику конфиденциальности](%s), нажав «Соглашаюсь»",
			publicOfferURL,
			privacyPolicyURL,
		),
		VerificationRequestSentToast:       "Отправлено пользователю",
		VerificationRequestReceivedToast:   "Запрос получен",
		MediaSentToast:                     "Медиа отправлены",
		SearchTokenPromptText:              "Укажите токен.\nПример:\n/search ZW6T-HJTK-6WY2",
		SearchNotFoundText:                 "❌ Ничего не найдено по указанному токену",
		SearchShowMediaButtonTemplate:      "📎 Показать медиа (%d)",
		SearchMissingOrderText:             "❌ Ошибка: данные заказа отсутствуют",
		SearchDealHeader:                   "🧾 <b>Сделка</b>\n",
		SearchStatusLineTemplate:           "Статус: <b>%s</b>\n",
		SearchCreatedLineTemplate:          "Создан: %s\n",
		SearchUpdatedLineTemplate:          "Обновлён: %s\n",
		SearchGameHeader:                   "🎮 <b>Игра</b>\n",
		SearchGameNameLineTemplate:         "Название: <b>%s</b>\n",
		SearchGameTypeLineTemplate:         "Тип: <b>%s</b>\n",
		SearchUserHeader:                   "👤 <b>Пользователь</b>\n",
		SearchUserNameLineTemplate:         "Имя: %s\n",
		SearchUserChatIDLineTemplate:       "Chat ID: <code>%d</code>\n",
		SearchUserVerifiedYes:              "Верифицирован: ✅\n",
		SearchUserVerifiedNo:               "Верифицирован: ❌\n",
		SearchUserTotalOrdersLineTemplate:  "Всего заказов: %d\n",
		SearchExpertHeader:                 "🧑‍💼 <b>Эксперт</b>\n",
		SearchExpertChatIDLineTemplate:     "Chat ID: <code>%d</code>\n",
		SearchExpertActiveYes:              "Активен: ✅\n",
		SearchExpertActiveNo:               "Активен: ❌\n",
		SearchUserStateHeader:              "📝 <b>Состояние пользователя</b>\n",
		SearchUserStateLineTemplate:        "State: <b>%s</b>\n",
		SearchUserStateUpdatedLineTemplate: "Обновлено: %s\n",
		SearchChatHeader:                   "\n💬 <b>Чат</b>\n",
		SenderUserLabel:                    "👤 Пользователь",
		SenderExpertLabel:                  "🧑‍💼 Эксперт",
		SenderSystemLabel:                  "⚙️ Система",
		ChatMessageHeaderTemplate:          "<b>%s</b> <i>%s</i>\n",
		ChatTextLineTemplate:               "\t\t\t\t\t\t> %s",
		ChatOtherLine:                      "\t\t\t\t\t\t> 🔡 <b>Другое</b>\n",
		ChatQuoteBlockTemplate:             "<blockquote expandable>\n%s\n</blockquote>",
		OrderStatusCreatedText:             "создан",
		OrderStatusAcceptedText:            "принят",
		OrderStatusExpertConfirmedText:     "подтверждён экспертом",
		OrderStatusCompletedText:           "подтверждён клиентом",
		OrderStatusDeclinedByExpertText:    "отменён экспертом",
		OrderStatusCanceledByUserText:      "отменён клиентом",
		UserStateIdleText:                  "в ожидании",
		UserStateStartText:                 "начало",
		UserStateCommunicationText:         "общается с экспертом",
		UserStateWritingReviewText:         "пишет отзыв",
		MediaPhotoLabel:                    "🖼 <b>Фото</b>\n",
		MediaVideoLabel:                    "🎥 <b>Видео</b>\n",
		MediaVideoNoteLabel:                "📹 <b>Кружок</b>\n",
		MediaDocumentWithNameTemplate:      "📎 <b>Документ</b> : %s\n",
		MediaDocumentLabel:                 "📎 <b>Документ</b>\n",
		MediaVoiceLabel:                    "🎤 <b>Голосовое сообщение</b>\n",
	}
}

func NewDynamic(privacyPolicyURL, publicOfferURL string) *Dynamic {
	return &Dynamic{
		privacyPolicyURL: privacyPolicyURL,
		publicOfferURL:   publicOfferURL,
	}
}

func (d *Dynamic) YouHaveChosenGame(gameName string) string {
	return fmt.Sprintf("_%s_", gameName)
}

func (d *Dynamic) YouHaveChosenGameAndType(g, t string) string {
	return fmt.Sprintf("_%s, %s_", g, t)
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
		"Эксперт принял твою заявку!\n\nТокен для обращения в поддержку:\n\n`%s` 🎉\n\nДальнейшее общение будет прямо здесь - удобно и быстро 😌",
		token,
	)
}

func (d *Dynamic) TitleOrderTopic(orderID int, itemGame, itemType string) string {
	return fmt.Sprintf("💼 Сделка #%d - (%s, %s)", orderID, itemGame, itemType)
}

func (d *Dynamic) ApplicationManagementText(
	gameName,
	gameTypeName string,
	isVerified bool,
) string {
	status := "не пройдена ❌"
	if isVerified {
		status = "пройдена ✅"
	}

	return fmt.Sprintf(
		"Панель управления заявкой ⚙️\n\nИгра: %s\nТип: %s\nВерификация: %s",
		gameName,
		gameTypeName,
		status,
	)
}

func (d *Dynamic) HelloText() string {
	return fmt.Sprintf(
		"👋Привет! Я Скупыч - бот, который превратит твой игровой опыт в реальные деньги💰\n\n"+
			"Моя задача - сделать процесс понятным и безопасным!\n\n"+
			"📑 Нажимая кнопку «Согласиться», ты подтверждаешь согласие на обработку персональных данных в соответствии с [Политикой конфиденциальности](%s) и принимаешь условия [Публичной оферты](%s).",
		d.privacyPolicyURL,
		d.publicOfferURL,
	)
}

func (d *Dynamic) HelloTextNotFirst() string {
	return fmt.Sprint(
		"👋Привет! Я Скупыч - бот, который превратит твой игровой опыт в реальные деньги💰\n\n",
		"Моя задача - сделать процесс понятным и безопасным!\n\n",
	)
}
