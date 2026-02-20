package domain

import (
	"context"
	"encoding/json"
	"time"
)

type CallbackToken struct {
	Token     string
	Action    string
	Payload   json.RawMessage
	CreatedAt time.Time
}

type CallbackTokenRepository interface {
	Create(ctx context.Context, callback *CallbackToken) error
	Consume(ctx context.Context, callback *CallbackToken) error
	Delete(ctx context.Context, token string, action string) error
	DeleteByActionAndOrderID(
		ctx context.Context,
		action string,
		orderID int,
	) error
}
