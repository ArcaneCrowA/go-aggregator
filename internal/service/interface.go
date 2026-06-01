package service

import (
	"context"

	"github.com/ArcaneCrowA/go-aggregator/internal/models"
)

type repo interface {
	AddSubscription(ctx context.Context, m models.SubscriptionModel) error
	GetSubscriptions(ctx context.Context, m models.CalculateQuery) ([]int, error)
}
