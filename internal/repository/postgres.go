package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"effective-mobile-task/internal/models"
)

// SubscriptionRepositoryInterface — контракт слоя repository.
// Определён здесь же, где и реализация, чтобы сервисный слой мог зависеть от абстракции.
type SubscriptionRepositoryInterface interface {
	Create(ctx context.Context, sub *models.Subscription) error
	GetByID(ctx context.Context, id string) (*models.Subscription, error)
	Update(ctx context.Context, id string, req *models.UpdateSubscriptionRequest) (*models.Subscription, error)
	Delete(ctx context.Context, id string) error
	List(ctx context.Context, limit, offset int, userID, serviceName string) ([]models.Subscription, error)
	GetTotalCost(ctx context.Context, userID, serviceName string, startDate, endDate time.Time) (int, error)
}

type SubscriptionRepository struct {
	db *sql.DB
}

func NewSubscriptionRepository(db *sql.DB) *SubscriptionRepository {
	return &SubscriptionRepository{db: db}
}

func (r *SubscriptionRepository) Create(ctx context.Context, sub *models.Subscription) error {
	query := `
		INSERT INTO subscriptions (service_name, price, user_id, start_date, end_date)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, created_at, updated_at
	`
	err := r.db.QueryRowContext(
		ctx, query, sub.ServiceName, sub.Price, sub.UserID, sub.StartDate, sub.EndDate,
	).Scan(&sub.ID, &sub.CreatedAt, &sub.UpdatedAt)

	return err
}

func (r *SubscriptionRepository) GetByID(ctx context.Context, id string) (*models.Subscription, error) {
	query := `
		SELECT id, service_name, price, user_id, start_date, end_date, created_at, updated_at
		FROM subscriptions
		WHERE id = $1
	`
	var sub models.Subscription
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&sub.ID, &sub.ServiceName, &sub.Price, &sub.UserID,
		&sub.StartDate, &sub.EndDate, &sub.CreatedAt, &sub.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, sql.ErrNoRows
	}
	return &sub, err
}

func (r *SubscriptionRepository) Update(ctx context.Context, id string, req *models.UpdateSubscriptionRequest) (*models.Subscription, error) {
	query := "UPDATE subscriptions SET updated_at = CURRENT_TIMESTAMP"
	args := []interface{}{}
	argCounter := 1

	if req.ServiceName != nil {
		query += fmt.Sprintf(", service_name = $%d", argCounter)
		args = append(args, *req.ServiceName)
		argCounter++
	}
	if req.Price != nil {
		query += fmt.Sprintf(", price = $%d", argCounter)
		args = append(args, *req.Price)
		argCounter++
	}
	if req.StartDate != nil {
		query += fmt.Sprintf(", start_date = $%d", argCounter)
		args = append(args, *req.StartDate)
		argCounter++
	}
	if req.EndDate != nil {
		query += fmt.Sprintf(", end_date = $%d", argCounter)
		args = append(args, *req.EndDate)
		argCounter++
	}

	query += fmt.Sprintf(" WHERE id = $%d RETURNING id, service_name, price, user_id, start_date, end_date, created_at, updated_at", argCounter)
	args = append(args, id)

	var sub models.Subscription
	err := r.db.QueryRowContext(ctx, query, args...).Scan(
		&sub.ID, &sub.ServiceName, &sub.Price, &sub.UserID,
		&sub.StartDate, &sub.EndDate, &sub.CreatedAt, &sub.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, sql.ErrNoRows
	}
	return &sub, err
}

func (r *SubscriptionRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM subscriptions WHERE id = $1`
	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rows == 0 {
		return sql.ErrNoRows
	}

	return nil
}

func (r *SubscriptionRepository) List(ctx context.Context, limit, offset int, userID, serviceName string) ([]models.Subscription, error) {
	query := `
		SELECT id, service_name, price, user_id, start_date, end_date, created_at, updated_at
		FROM subscriptions
		WHERE 1=1
	`
	args := []interface{}{}
	argCounter := 1

	if userID != "" {
		query += fmt.Sprintf(" AND user_id = $%d", argCounter)
		args = append(args, userID)
		argCounter++
	}
	if serviceName != "" {
		query += fmt.Sprintf(" AND service_name = $%d", argCounter)
		args = append(args, serviceName)
		argCounter++
	}

	query += " ORDER BY created_at DESC"
	query += fmt.Sprintf(" LIMIT $%d OFFSET $%d", argCounter, argCounter+1)
	args = append(args, limit, offset)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var subscriptions []models.Subscription
	for rows.Next() {
		var sub models.Subscription
		err := rows.Scan(
			&sub.ID, &sub.ServiceName, &sub.Price, &sub.UserID,
			&sub.StartDate, &sub.EndDate, &sub.CreatedAt, &sub.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		subscriptions = append(subscriptions, sub)
	}

	return subscriptions, nil
}

func (r *SubscriptionRepository) GetTotalCost(ctx context.Context, userID, serviceName string, startDate, endDate time.Time) (int, error) {
	query := `
		SELECT COALESCE(SUM(
			price * GREATEST(1, CEIL(
				(
					LEAST(COALESCE(end_date, $4::date), $4::date)
					- GREATEST(start_date, $3::date)
				)::numeric / 30.0
			))
		), 0)
		FROM subscriptions
		WHERE user_id = $1
			AND service_name = $2
			AND start_date <= $4
			AND (end_date IS NULL OR end_date >= $3)
	`
	var total int
	err := r.db.QueryRowContext(ctx, query, userID, serviceName, startDate, endDate).Scan(&total)
	return total, err
}
