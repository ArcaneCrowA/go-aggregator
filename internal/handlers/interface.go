package handlers

import (
	"context"

	"github.com/ArcaneCrowA/go-aggregator/internal/models"
)

type Service interface {
	AddSubscription(context.Context, models.AddSubscriptionDTO) error
	Calculate(context.Context, models.CalculateQuery) (int, error)
}
