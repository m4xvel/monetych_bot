package postgres

import (
	"context"
	"database/sql"
	"fmt"

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

func (r *OrderRepo) Create(ctx context.Context, userID int) (int, error) {
	const q = `
	INSERT INTO orders (user_id, status, created_at, updated_at)
	VALUES ($1, $2, now(), now())
	RETURNING id`
	var id int
	if err := r.pool.QueryRow(ctx, q, userID, domain.OrderNew).Scan(&id); err != nil {
		return 0, fmt.Errorf("order repo - Create: %w", err)
	}
	return id, nil
}

func (r *OrderRepo) Get(ctx context.Context, id int) (*domain.Order, error) {
	const q = `
	SELECT id, user_id, appraiser_id, status, topic_id, thread_id
	FROM orders
	WHERE id = $1 LIMIT 1`
	var o domain.Order
	var appraiser sql.NullInt32
	var topic sql.NullInt64
	var thread sql.NullInt64

	err := r.pool.QueryRow(ctx, q, id).Scan(
		&o.ID,
		&o.UserID,
		&appraiser,
		&o.Status,
		&topic,
		&thread,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("order repo - Get: %w", err)
	}

	if appraiser.Valid {
		val := int(appraiser.Int32)
		o.AppraiserID = &val
	} else {
		o.AppraiserID = nil
	}
	if topic.Valid {
		val := topic.Int64
		o.TopicID = &val
	} else {
		o.TopicID = nil
	}
	if thread.Valid {
		val := thread.Int64
		o.ThreadID = &val
	} else {
		o.ThreadID = nil
	}

	return &o, nil
}

func (r *OrderRepo) GetByUser(ctx context.Context, userID int, status domain.OrderStatus) (*domain.Order, error) {
	const q = `
	SELECT id, user_id, appraiser_id, status, topic_id, thread_id
	FROM orders
	WHERE user_id = $1 AND status = $2
	LIMIT 1`
	var o domain.Order
	var appraiser sql.NullInt32
	var topic sql.NullInt64
	var thread sql.NullInt64

	err := r.pool.QueryRow(ctx, q, userID, status).Scan(
		&o.ID,
		&o.UserID,
		&appraiser,
		&o.Status,
		&topic,
		&thread,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("order repo - GetByUser: %w", err)
	}

	if appraiser.Valid {
		val := int(appraiser.Int32)
		o.AppraiserID = &val
	}
	if topic.Valid {
		val := topic.Int64
		o.TopicID = &val
	}
	if thread.Valid {
		val := thread.Int64
		o.ThreadID = &val
	}
	return &o, nil
}

func (r *OrderRepo) GetByThread(ctx context.Context, topicID, threadID int64) (*domain.Order, error) {
	const q = `
	SELECT id, user_id, appraiser_id, status, topic_id, thread_id
	FROM orders
	WHERE topic_id = $1 AND thread_id = $2 AND status = 'active'
	LIMIT 1`
	var o domain.Order
	var appraiser sql.NullInt32
	var topic sql.NullInt64
	var thread sql.NullInt64

	err := r.pool.QueryRow(ctx, q, topicID, threadID).Scan(
		&o.ID,
		&o.UserID,
		&appraiser,
		&o.Status,
		&topic,
		&thread,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("order repo - GetByThread: %w", err)
	}

	if appraiser.Valid {
		val := int(appraiser.Int32)
		o.AppraiserID = &val
	}
	if topic.Valid {
		val := topic.Int64
		o.TopicID = &val
	}
	if thread.Valid {
		val := thread.Int64
		o.ThreadID = &val
	}
	return &o, nil
}

func (r *OrderRepo) Accept(ctx context.Context, orderID int, assessorID int, topicID, threadID int64) (*domain.Order, error) {
	const q = `
	UPDATE orders
	SET appraiser_id = $1,
			topic_id = $2,
			thread_id = $3,
			status = 'active',
			updated_at = NOW()
	WHERE id = $4 AND status = 'new'
	RETURNING id, user_id, appraiser_id, status, topic_id, thread_id`

	var o domain.Order
	var appraiserID *int
	var topicIDNullable, threadIDNullable *int64

	err := r.pool.QueryRow(ctx, q, assessorID, topicID, threadID, orderID).Scan(
		&o.ID,
		&o.UserID,
		&appraiserID,
		&o.Status,
		&topicIDNullable,
		&threadIDNullable,
	)
	if err != nil {
		return nil, fmt.Errorf("accept order: %w", err)
	}

	o.AppraiserID = appraiserID
	o.TopicID = topicIDNullable
	o.ThreadID = threadIDNullable

	return &o, nil
}

func (r *OrderRepo) AssignAssessor(ctx context.Context, orderID, assessorID int) error {
	const q = `
	UPDATE orders
	SET appraiser_id = $1, updated_at = now()
	WHERE id = $2`
	ct, err := r.pool.Exec(ctx, q, assessorID, orderID)
	if err != nil {
		return fmt.Errorf("order repo - AssignAssessor: %w", err)
	}
	if ct.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *OrderRepo) SetThread(ctx context.Context, orderID int, topicID, threadID int64) error {
	const q = `
	UPDATE orders
	SET topic_id = $1, thread_id = $2, updated_at = now()
	WHERE id = $3`
	ct, err := r.pool.Exec(ctx, q, topicID, threadID, orderID)
	if err != nil {
		return fmt.Errorf("order repo - SetThread: %w", err)
	}
	if ct.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *OrderRepo) UpdateStatus(ctx context.Context, orderID int, status domain.OrderStatus) error {
	const q = `
	UPDATE orders
	SET status = $1, updated_at = now()
	WHERE id = $2`
	_, err := r.pool.Exec(ctx, q, status, orderID)
	if err != nil {
		return fmt.Errorf("order repo - UpdateStatus: %w", err)
	}
	return nil
}
