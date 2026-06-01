package handlers

import (
	"net/http"
	"time"

	"github.com/ArcaneCrowA/go-aggregator/internal/models"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type handler struct {
	log *zap.Logger
	svc Service
}

func New(log *zap.Logger, svc Service) *handler {
	return &handler{
		log: log,
		svc: svc,
	}
}

func (h *handler) AddSubscription(c *gin.Context) {
	var dto models.AddSubscriptionDTO
	if err := c.ShouldBindJSON(&dto); err != nil {
		h.log.Error("failed parsing", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := dto.Validate(); err != nil {
		h.log.Error("failed validation", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.svc.AddSubscription(c.Request.Context(), dto); err != nil {
		h.log.Error("failed validation", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusAccepted, gin.H{"status": "accepted"})
}

func (h *handler) GetSubscriptionsFilter(c *gin.Context) {
	var data models.CalculateQuery

	data.Name = c.Query("service_name")
	userIDRaw := c.Query("user_id")
	if err := uuid.Validate(userIDRaw); err != nil {
		h.log.Error("failed validation", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "incorrect uuid user_id"})
		return
	}
	data.ClientID = uuid.MustParse(userIDRaw)

	parse := func(t string) *time.Time {
		if t == "" {
			return nil
		}
		parsed, err := time.Parse("01-2006", t)
		if err != nil {
			return nil
		}
		return &parsed
	}
	startDateRaw := c.Query("start_date")
	endDateRaw := c.Query("end_date")

	data.StartDate = parse(startDateRaw)
	data.EndDate = parse(endDateRaw)

	value, err := h.svc.Calculate(c.Request.Context(), data)
	if err != nil {
		h.log.Error("failed to calculate", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "accepted", "total": value})
}
