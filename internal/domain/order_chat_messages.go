package domain

import (
	"context"
	"time"
)

type SenderRole string
type MessageType string

const (
	SenderUser   SenderRole = "user"
	SenderExpert SenderRole = "expert"
	SenderSystem SenderRole = "system"

	MessageText     MessageType = "text"
	MessagePhoto    MessageType = "photo"
	MessageVideo    MessageType = "video"
	MessageDocument MessageType = "document"
	MessageVoice    MessageType = "voice"
	MessageOther    MessageType = "other"
)

type OrderChatMessages struct {
	ID             int64
	OrderID        int
	SenderRole     SenderRole
	SenderUserID   *int
	SenderExpertID *int
	ChatID         int64
	MessageID      int
	MessageType    MessageType
	Text           *string
	Media          map[string]any
	RawPayload     map[string]any
	CreatedAt      time.Time
}

type OrderChatMessagesRepository interface {
	Save(ctx context.Context, msg *OrderChatMessages) error
}
