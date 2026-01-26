package postgres

import (
	"context"
	"encoding/json"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/m4xvel/monetych_bot/internal/crypto"
	"github.com/m4xvel/monetych_bot/internal/domain"
	"github.com/m4xvel/monetych_bot/internal/logger"
)

type OrderChatMessagesRepo struct {
	pool   *pgxpool.Pool
	crypto *crypto.Service
}

func NewOrderChatMessagesRepo(
	pool *pgxpool.Pool,
	crypto *crypto.Service,
) *OrderChatMessagesRepo {
	return &OrderChatMessagesRepo{
		pool:   pool,
		crypto: crypto,
	}
}

func (r *OrderChatMessagesRepo) Save(
	ctx context.Context,
	msg *domain.OrderChatMessages,
) error {
	var textEnc []byte
	var mediaEnc []byte
	var err error

	if msg.Text != nil {
		textEnc, err = r.crypto.Encrypt([]byte(*msg.Text))
		if err != nil {
			return err
		}
	}

	if msg.Media != nil {
		raw, err := json.Marshal(msg.Media)
		if err != nil {
			return err
		}

		mediaEnc, err = r.crypto.Encrypt(raw)
		if err != nil {
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
			text_enc,
			media_enc
		)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)
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
		textEnc,
		mediaEnc,
	)

	if err != nil {
		logger.Log.Errorw("insert chat message failed", "err", err)
		return err
	}

	if cmd.RowsAffected() == 0 {
		return nil
	}

	return nil
}
