package postgres

import (
	"context"
	"encoding/json"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/m4xvel/monetych_bot/internal/domain"
	"github.com/m4xvel/monetych_bot/internal/logger"
)

type OrderChatMessagesRepo struct {
	pool *pgxpool.Pool
}

func NewOrderChatMessagesRepo(pool *pgxpool.Pool) *OrderChatMessagesRepo {
	return &OrderChatMessagesRepo{pool: pool}
}

func (r *OrderChatMessagesRepo) Save(
	ctx context.Context,
	msg *domain.OrderChatMessages,
) error {
	var mediaJSON []byte
	var rawJSON []byte
	var err error

	if msg.Media != nil {
		mediaJSON, err = json.Marshal(msg.Media)
		if err != nil {
			logger.Log.Errorw("failed to marshaling msg.Media",
				"err", err,
			)
			return err
		}
	}

	if msg.RawPayload != nil {
		rawJSON, err = json.Marshal(msg.RawPayload)
		if err != nil {
			logger.Log.Errorw("failed to marshaling msg.RawPayload",
				"err", err,
			)
			return err
		}
	}

	const q = `
		INSERT INTO order_chat_messages (
			order_id,
			sender_role,
			sender_user_id,
			sender_expert_id,
			chat_id,
			message_id,
			message_type,
			text,
			media,
			raw_payload
		)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)
		ON CONFLICT DO NOTHING
	`

	cmd, err := r.pool.Exec(ctx, q,
		msg.OrderID,
		msg.SenderRole,
		msg.SenderUserID,
		msg.SenderExpertID,
		msg.ChatID,
		msg.MessageID,
		msg.MessageType,
		msg.Text,
		mediaJSON,
		rawJSON,
	)

	if err != nil {
		logger.Log.Errorw("failed to insert chat_message",
			"err", err,
		)
		return err
	}

	if cmd.RowsAffected() == 0 {
		return nil
	}

	return nil
}
