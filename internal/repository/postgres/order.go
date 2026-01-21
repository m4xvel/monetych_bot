package postgres

import (
	"context"
	"database/sql"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/m4xvel/monetych_bot/internal/domain"
)

type OrderRepo struct {
	pool *pgxpool.Pool
}

func NewOrderRepo(pool *pgxpool.Pool) *OrderRepo {
	return &OrderRepo{pool: pool}
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

	if err == pgx.ErrNoRows {
		return 0, nil
	}

	if err != nil {
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
		return nil, err
	}

	return &o, nil
}

func (r *OrderRepo) FindByToken(
	ctx context.Context,
	token string,
) (*domain.OrderFull, error) {
	const q = `
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
			WHERE o.order_token = $1
		`

	of := domain.OrderFull{
		User:      &domain.User{},
		Expert:    &domain.Expert{},
		Game:      &domain.Game{},
		GameType:  &domain.GameType{},
		UserState: &domain.UserState{},
	}
	err := r.pool.QueryRow(ctx, q, token).Scan(
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

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
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

	_ = r.pool.QueryRow(ctx, userStateQ, of.User.ID).Scan(
		&of.UserState.State,
		&of.UserState.OrderID,
		&of.UserState.UpdatedAt,
	)

	return &of, nil
}
