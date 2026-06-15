package handlers

import (
	"database/sql"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"effective-mobile-task/internal/models"
	"effective-mobile-task/internal/service"
)

// @Description API for managing subscriptions.

const (
	defaultLimit = 20
	maxLimit     = 100
)

type SubscriptionHandler struct {
	service service.SubscriptionServiceInterface
}

func NewSubscriptionHandler(s service.SubscriptionServiceInterface) *SubscriptionHandler {
	return &SubscriptionHandler{service: s}
}

// CreateSubscription
// @Summary Create a new subscription
// @Description Creates a new subscription record
// @Tags Subscriptions
// @Accept json
// @Produce json
// @Param request body models.CreateSubscriptionRequest true "Subscription data"
// @Success 201 {object} models.Subscription
// @Failure 400 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /subscriptions [post]
func (h *SubscriptionHandler) CreateSubscription(c *gin.Context) {
	var req models.CreateSubscriptionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	sub := &models.Subscription{
		ServiceName: req.ServiceName,
		Price:       req.Price,
		UserID:      req.UserID,
		StartDate:   req.StartDate,
		EndDate:     req.EndDate,
	}

	if err := h.service.Create(c.Request.Context(), sub); err != nil {
		var valErr *service.ValidationError
		if errors.As(err, &valErr) {
			c.JSON(http.StatusBadRequest, gin.H{"error": valErr.Message})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	c.JSON(http.StatusCreated, sub)
}

// GetSubscription
// @Summary Get a subscription by ID
// @Description Returns a single subscription by its UUID
// @Tags Subscriptions
// @Produce json
// @Param id path string true "Subscription ID (UUID)"
// @Success 200 {object} models.Subscription
// @Failure 400 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /subscriptions/{id} [get]
func (h *SubscriptionHandler) GetSubscription(c *gin.Context) {
	id := c.Param("id")
	if _, err := uuid.Parse(id); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id format"})
		return
	}

	sub, err := h.service.GetByID(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			c.JSON(http.StatusNotFound, gin.H{"error": "record not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	c.JSON(http.StatusOK, sub)
}

// UpdateSubscription
// @Summary Update a subscription
// @Description Updates an existing subscription by ID (partial update)
// @Tags Subscriptions
// @Accept json
// @Produce json
// @Param id path string true "Subscription ID (UUID)"
// @Param request body models.UpdateSubscriptionRequest true "Subscription fields to update"
// @Success 200 {object} models.Subscription
// @Failure 400 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /subscriptions/{id} [put]
func (h *SubscriptionHandler) UpdateSubscription(c *gin.Context) {
	id := c.Param("id")
	var req models.UpdateSubscriptionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	sub, err := h.service.Update(c.Request.Context(), id, &req)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			c.JSON(http.StatusNotFound, gin.H{"error": "record not found"})
			return
		}
		var valErr *service.ValidationError
		if errors.As(err, &valErr) {
			c.JSON(http.StatusBadRequest, gin.H{"error": valErr.Message})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	c.JSON(http.StatusOK, sub)
}

// DeleteSubscription
// @Summary Delete a subscription
// @Description Deletes a subscription by ID
// @Tags Subscriptions
// @Param id path string true "Subscription ID (UUID)"
// @Success 204 "No Content"
// @Failure 404 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /subscriptions/{id} [delete]
func (h *SubscriptionHandler) DeleteSubscription(c *gin.Context) {
	id := c.Param("id")
	if err := h.service.Delete(c.Request.Context(), id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			c.JSON(http.StatusNotFound, gin.H{"error": "record not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	c.Status(http.StatusNoContent)
}

// ListSubscriptions
// @Summary List subscriptions
// @Description Returns a paginated list of subscriptions, optionally filtered by user_id and/or service_name
// @Tags Subscriptions
// @Produce json
// @Param limit query int false "Max results (default 20, max 100)"
// @Param offset query int false "Offset for pagination"
// @Param user_id query string false "Filter by user ID (UUID)"
// @Param service_name query string false "Filter by service name"
// @Success 200 {array} models.Subscription
// @Failure 500 {object} models.ErrorResponse
// @Router /subscriptions [get]
func (h *SubscriptionHandler) ListSubscriptions(c *gin.Context) {
	limit := defaultLimit
	offset := 0

	if l, err := strconv.Atoi(c.Query("limit")); err == nil && l > 0 {
		if l > maxLimit {
			limit = maxLimit
		} else {
			limit = l
		}
	}

	if o, err := strconv.Atoi(c.Query("offset")); err == nil && o >= 0 {
		offset = o
	}

	userID := c.Query("user_id")
	serviceName := c.Query("service_name")

	subscriptions, err := h.service.List(c.Request.Context(), limit, offset, userID, serviceName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	c.JSON(http.StatusOK, subscriptions)
}

// GetTotalCost
// @Summary Calculate total cost
// @Description Returns the total cost of subscriptions for a user and service within a date range
// @Tags Subscriptions
// @Produce json
// @Param user_id query string true "User ID (UUID)"
// @Param service_name query string true "Service name"
// @Param start_date query string true "Start date (YYYY-MM-DD)"
// @Param end_date query string true "End date (YYYY-MM-DD)"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /subscriptions/total-cost [get]
func (h *SubscriptionHandler) GetTotalCost(c *gin.Context) {
	userID := c.Query("user_id")
	serviceName := c.Query("service_name")
	startDateStr := c.Query("start_date")
	endDateStr := c.Query("end_date")

	if userID == "" || serviceName == "" || startDateStr == "" || endDateStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing required parameters"})
		return
	}

	if _, err := uuid.Parse(userID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user_id format"})
		return
	}

	startDate, err := time.Parse("2006-01-02", startDateStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid start_date format, use YYYY-MM-DD"})
		return
	}

	endDate, err := time.Parse("2006-01-02", endDateStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid end_date format, use YYYY-MM-DD"})
		return
	}

	total, err := h.service.GetTotalCost(c.Request.Context(), userID, serviceName, startDate, endDate)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"total_cost": total})
}
