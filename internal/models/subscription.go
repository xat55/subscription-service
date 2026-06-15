package models

import (
	"time"
)

type Subscription struct {
	ID          string     `json:"id"`
	ServiceName string     `json:"service_name"`
	Price       int        `json:"price"`
	UserID      string     `json:"user_id"`
	StartDate   time.Time  `json:"start_date" example:"2024-01-01T00:00:00Z"`
	EndDate     *time.Time `json:"end_date,omitempty" example:"2024-12-31T00:00:00Z"`
	CreatedAt   time.Time  `json:"created_at" example:"2024-01-01T00:00:00Z"`
	UpdatedAt   time.Time  `json:"updated_at" example:"2024-01-01T00:00:00Z"`
}

type CreateSubscriptionRequest struct {
	ServiceName string     `json:"service_name" binding:"required"`
	Price       int        `json:"price" binding:"required,min=0"`
	UserID      string     `json:"user_id" binding:"required"`
	StartDate   time.Time  `json:"start_date" binding:"required" example:"2024-01-01T00:00:00Z"`
	EndDate     *time.Time `json:"end_date,omitempty" example:"2024-12-31T00:00:00Z"`
}

type UpdateSubscriptionRequest struct {
	ServiceName *string    `json:"service_name,omitempty"`
	Price       *int       `json:"price,omitempty"`
	StartDate   *time.Time `json:"start_date,omitempty" example:"2024-01-01T00:00:00Z"`
	EndDate     *time.Time `json:"end_date,omitempty" example:"2024-12-31T00:00:00Z"`
}

type ErrorResponse struct {
	Error string `json:"error" example:"error message"`
}
