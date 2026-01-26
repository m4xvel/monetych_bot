package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/m4xvel/monetych_bot/internal/crypto"
	"github.com/m4xvel/monetych_bot/internal/domain"
	"github.com/m4xvel/monetych_bot/internal/logger"
)

type OrderRepo struct {
	pool   *pgxpool.Pool
	crypto *crypto.Service
}

func NewOrderRepo(
	pool *pgxpool.Pool,
	crypto *crypto.Service,
) *OrderRepo {
	return &OrderRepo{
		pool:   pool,
		crypto: crypto,
	}
}

func (r *OrderRepo) Create(ctx context.Context, order domain.Order) (int, error) {
	const q = `
	INSERT INTO orders (
		user_id, 
		game_id, 
		game_type_id, 
		user_name_at_purchase, 
		game_name_at_purchase, 
		game_type_name_at_purchase
	)
	VALUES ($1, $2, $3, $4, $5, $6)
	ON CONFLICT DO NOTHING
	RETURNING id
	`
	var id int
	err := r.pool.QueryRow(
		ctx,
		q,
		order.UserID,
		order.GameID,
		order.GameTypeID,
		order.UserNameAtPurchase,
		order.GameNameAtPurchase,
		order.GameTypeNameAtPurchase).Scan(&id)

	if err != nil && err != pgx.ErrNoRows {
		logger.Log.Errorw("order repo: create failed",
			"user_id", order.UserID,
			"err", err,
		)
		return 0, err
	}

	return id, nil
}

func (r *OrderRepo) UpdateStatus(
	ctx context.Context,
	order domain.Order,
	status domain.OrderStatus,
) error {
	const q = `
	UPDATE orders 
	SET 
		status = $2, 
		updated_at = now()
	WHERE id = $1
		AND status = $3
	`
	cmd, err := r.pool.Exec(ctx, q, order.ID, order.Status, status)

	if err != nil {
		logger.Log.Errorw("order repo: update status failed",
			"order_id", order.ID,
			"to_status", order.Status,
			"err", err,
		)
		return err
	}

	if cmd.RowsAffected() == 0 {
		return ErrOrderAlreadyProcessed
	}

	return nil
}

func (r *OrderRepo) SetActive(
	ctx context.Context,
	order domain.Order,
	status domain.OrderStatus,
) error {
	const q = `
	UPDATE orders 
	SET 
		expert_id = $2,
		thread_id = $3,
		updated_at = now()
	WHERE id = $1
		AND status = $4
	`
	cmd, err := r.pool.Exec(
		ctx, q,
		order.ID,
		order.ExpertID,
		order.ThreadID,
		status,
	)

	if err != nil {
		logger.Log.Errorw("order repo: set active failed",
			"order_id", order.ID,
			"expert_id", order.ExpertID,
			"err", err,
		)
		return err
	}

	if cmd.RowsAffected() == 0 {
		return ErrOrderAlreadyProcessed
	}

	return nil
}

func (r *OrderRepo) Get(ctx context.Context, orderID int) (*domain.Order, error) {
	const q = `
		SELECT 
			o.id,
			substr(o.order_token,1,4) || '-' ||
			substr(o.order_token,5,4) || '-' ||
			substr(o.order_token,9,4) AS pretty_token,
			o.thread_id,
			o.game_name_at_purchase,
			o.game_type_name_at_purchase,
			u.chat_id,
			e.topic_id
		FROM orders o
		LEFT JOIN users u ON u.id = o.user_id
		LEFT JOIN experts e ON e.id = o.expert_id
		WHERE o.id = $1
	`
	var o domain.Order
	if err := r.pool.QueryRow(ctx, q, orderID).Scan(
		&o.ID,
		&o.Token,
		&o.ThreadID,
		&o.GameNameAtPurchase,
		&o.GameTypeNameAtPurchase,
		&o.UserChatID,
		&o.TopicID,
	); err != nil {
		logger.Log.Errorw("order repo: get failed",
			"order_id", orderID,
			"err", err,
		)
		return nil, err
	}

	return &o, nil
}

func (r *OrderRepo) FindByField(
	ctx context.Context,
	where string,
	arg any,
) (*domain.OrderFull, error) {
	q := fmt.Sprintf(`
			SELECT
				o.id, 
				o.order_token,
				o.status, 
				o.thread_id, 
				o.created_at, 
				o.updated_at,
				o.user_name_at_purchase, 
				o.game_name_at_purchase, 
				o.game_type_name_at_purchase,
				u.id, 
				u.chat_id, 
				u.name, 
				u.is_verified, 
				u.created_at, 
				u.total_orders,
				e.id, 
				e.chat_id, 
				e.topic_id, 
				e.is_active,
				g.id, g.name,
				gt.id, gt.name
			FROM orders o
			LEFT JOIN users u ON u.id = o.user_id
			LEFT JOIN experts e ON e.id = o.expert_id
			LEFT JOIN games g ON g.id = o.game_id
			LEFT JOIN game_types gt ON gt.id = o.game_type_id
			WHERE %s
	`, where)

	of := domain.OrderFull{
		User:      &domain.User{},
		Expert:    &domain.Expert{},
		Game:      &domain.Game{},
		GameType:  &domain.GameType{},
		UserState: &domain.UserState{},
	}
	err := r.pool.QueryRow(ctx, q, arg).Scan(
		&of.Order.ID,
		&of.Order.Token,
		&of.Order.Status,
		&of.Order.ThreadID,
		&of.Order.CreatedAt,
		&of.Order.UpdatedAt,
		&of.Order.UserNameAtPurchase,
		&of.Order.GameNameAtPurchase,
		&of.Order.GameTypeNameAtPurchase,
		&of.User.ID,
		&of.User.ChatID,
		&of.User.Name,
		&of.User.IsVerified,
		&of.User.CreatedAt,
		&of.User.TotalOrders,
		&of.Expert.ID,
		&of.Expert.ChatID,
		&of.Expert.TopicID,
		&of.Expert.IsActive,
		&of.Game.ID,
		&of.Game.Name,
		&of.GameType.ID,
		&of.GameType.Name,
	)

	if err != nil && err != sql.ErrNoRows {
		logger.Log.Errorw("order repo: find by token failed",
			"err", err,
		)
		return nil, err
	}

	const userStateQ = `
		SELECT 
			state, 
			order_id, 
			updated_at
		FROM user_state
		WHERE user_id = $1
	`

	err = r.pool.QueryRow(ctx, userStateQ, of.User.ID).Scan(
		&of.UserState.State,
		&of.UserState.OrderID,
		&of.UserState.UpdatedAt,
	)

	if err != nil && err != sql.ErrNoRows {
		logger.Log.Warnw("order repo: user state not found",
			"user_id", of.User.ID,
		)
	}

	const messagesQ = `
		SELECT
			sender_role,
			message_type,
			text_enc,
    	media_enc,
			created_at
		FROM order_chat_messages
		WHERE order_id = $1
		ORDER BY created_at ASC
	`

	rows, err := r.pool.Query(ctx, messagesQ, of.Order.ID)
	if err != nil {
		logger.Log.Errorw("order repo: failed to load messages",
			"order_id", of.Order.ID,
			"err", err,
		)
		return &of, nil
	}
	defer rows.Close()

	for rows.Next() {
		var (
			msg      domain.ChatMessage
			textEnc  []byte
			mediaEnc []byte
		)

		if err := rows.Scan(
			&msg.SenderRole,
			&msg.MessageType,
			&textEnc,
			&mediaEnc,
			&msg.CreatedAt,
		); err != nil {
			logger.Log.Warnw("order repo: failed to scan message",
				"order_id", of.Order.ID,
				"err", err,
			)
			continue
		}

		if len(textEnc) > 0 {
			raw, err := r.crypto.Decrypt(textEnc)
			if err == nil {
				s := string(raw)
				msg.Text = &s
			}
		}

		if len(mediaEnc) > 0 {
			raw, err := r.crypto.Decrypt(mediaEnc)
			if err == nil {
				var m map[string]any
				if json.Unmarshal(raw, &m) == nil {
					msg.Media = m
				}
			}
		}

		of.Messages = append(of.Messages, msg)
	}

	return &of, nil
}

func (r *OrderRepo) FindByToken(ctx context.Context, token string) (*domain.OrderFull, error) {
	return r.FindByField(ctx, "o.order_token = $1", token)
}

func (r *OrderRepo) FindByID(ctx context.Context, id int) (*domain.OrderFull, error) {
	return r.FindByField(ctx, "o.id = $1", id)
}
