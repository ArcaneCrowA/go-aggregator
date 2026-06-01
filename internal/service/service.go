package service

import (
	"context"
	"time"

	"github.com/ArcaneCrowA/go-aggregator/internal/models"
)

type service struct {
	repo repo
}

func New(repo repo) *service {
	return &service{
		repo: repo,
	}
}

func (s *service) AddSubscription(ctx context.Context, m models.AddSubscriptionDTO) error {
	startDate, err := time.Parse("01-2006", m.StartDate)
	if err != nil {
		return err
	}

	data := models.SubscriptionModel{
		Name:      m.Name,
		Price:     *m.Price,
		UserID:    m.UserID,
		StartDate: startDate,
	}

	if m.EndDate != nil && *m.EndDate != "" {
		endDate, err := time.Parse("01-2006", *m.EndDate)
		if err != nil {
			return err
		}
		data.EndDate = &endDate
	}

	return s.repo.AddSubscription(ctx, data)
}

func (s *service) Calculate(ctx context.Context, m models.CalculateQuery) (int, error) {
	prices, err := s.repo.GetSubscriptions(ctx, m)
	if err != nil {
		return 0, err
	}

	var sum int
	for _, price := range prices {
		sum += price
	}
	return sum, nil
}
