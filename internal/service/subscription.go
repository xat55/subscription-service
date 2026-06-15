package service

import (
	"context"
	"database/sql"
	"errors"
	"log/slog"
	"time"

	"effective-mobile-task/internal/models"
	"effective-mobile-task/internal/repository"
)

type ValidationError struct {
	Message string
}

func (e *ValidationError) Error() string {
	return e.Message
}

type SubscriptionServiceInterface interface {
	Create(ctx context.Context, sub *models.Subscription) error
	GetByID(ctx context.Context, id string) (*models.Subscription, error)
	Update(ctx context.Context, id string, req *models.UpdateSubscriptionRequest) (*models.Subscription, error)
	Delete(ctx context.Context, id string) error
	List(ctx context.Context, limit, offset int, userID, serviceName string) ([]models.Subscription, error)
	GetTotalCost(ctx context.Context, userID, serviceName string, startDate, endDate time.Time) (int, error)
}

type SubscriptionService struct {
	repo repository.SubscriptionRepositoryInterface
	log  *slog.Logger
}

func NewSubscriptionService(repo repository.SubscriptionRepositoryInterface, log *slog.Logger) *SubscriptionService {
	return &SubscriptionService{repo: repo, log: log}
}

func (s *SubscriptionService) Create(ctx context.Context, sub *models.Subscription) error {
	if sub.Price < 0 {
		s.log.Warn("validation failed: negative price",
			"user_id", sub.UserID,
			"service_name", sub.ServiceName,
			"price", sub.Price,
		)
		return &ValidationError{Message: "price cannot be negative"}
	}

	if sub.EndDate != nil && sub.EndDate.Before(sub.StartDate) {
		s.log.Warn("validation failed: end_date before start_date",
			"user_id", sub.UserID,
			"service_name", sub.ServiceName,
			"start_date", sub.StartDate,
			"end_date", *sub.EndDate,
		)
		return &ValidationError{Message: "end date cannot be before start date"}
	}

	if err := s.repo.Create(ctx, sub); err != nil {
		s.log.Error("failed to create subscription",
			"user_id", sub.UserID,
			"service_name", sub.ServiceName,
			"error", err,
		)
		return err
	}

	s.log.Info("subscription created",
		"id", sub.ID,
		"user_id", sub.UserID,
		"service_name", sub.ServiceName,
		"price", sub.Price,
	)

	return nil
}

func (s *SubscriptionService) GetByID(ctx context.Context, id string) (*models.Subscription, error) {
	sub, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			s.log.Warn("subscription not found", "id", id)
		} else {
			s.log.Error("failed to get subscription", "id", id, "error", err)
		}
		return nil, err
	}

	s.log.Debug("subscription retrieved", "id", id, "user_id", sub.UserID)
	return sub, nil
}

func (s *SubscriptionService) Update(ctx context.Context, id string, req *models.UpdateSubscriptionRequest) (*models.Subscription, error) {
	sub, err := s.repo.Update(ctx, id, req)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			s.log.Warn("subscription not found for update", "id", id)
		} else {
			s.log.Error("failed to update subscription", "id", id, "error", err)
		}
		return nil, err
	}

	s.log.Info("subscription updated",
		"id", id,
		"user_id", sub.UserID,
		"service_name", sub.ServiceName,
	)

	return sub, nil
}

func (s *SubscriptionService) Delete(ctx context.Context, id string) error {
	if err := s.repo.Delete(ctx, id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			s.log.Warn("subscription not found for deletion", "id", id)
		} else {
			s.log.Error("failed to delete subscription", "id", id, "error", err)
		}
		return err
	}

	s.log.Info("subscription deleted", "id", id)
	return nil
}

func (s *SubscriptionService) List(ctx context.Context, limit, offset int, userID, serviceName string) ([]models.Subscription, error) {
	subscriptions, err := s.repo.List(ctx, limit, offset, userID, serviceName)
	if err != nil {
		s.log.Error("failed to list subscriptions",
			"limit", limit,
			"offset", offset,
			"user_id", userID,
			"service_name", serviceName,
			"error", err,
		)
		return nil, err
	}

	s.log.Debug("subscriptions listed",
		"count", len(subscriptions),
		"user_id", userID,
		"service_name", serviceName,
	)
	return subscriptions, nil
}

func (s *SubscriptionService) GetTotalCost(ctx context.Context, userID, serviceName string, startDate, endDate time.Time) (int, error) {
	total, err := s.repo.GetTotalCost(ctx, userID, serviceName, startDate, endDate)
	if err != nil {
		s.log.Error("failed to calculate total cost",
			"user_id", userID,
			"service_name", serviceName,
			"start_date", startDate,
			"end_date", endDate,
			"error", err,
		)
		return 0, err
	}

	s.log.Info("total cost calculated",
		"user_id", userID,
		"service_name", serviceName,
		"start_date", startDate,
		"end_date", endDate,
		"total_cost", total,
	)

	return total, nil
}
