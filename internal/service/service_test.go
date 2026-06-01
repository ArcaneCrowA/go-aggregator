package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/ArcaneCrowA/go-aggregator/internal/models"
	"github.com/google/uuid"
)

type mockRepo struct {
	addFn func(ctx context.Context, m models.SubscriptionModel) error
	getFn func(ctx context.Context, m models.CalculateQuery) ([]int, error)
}

func (m *mockRepo) AddSubscription(ctx context.Context, s models.SubscriptionModel) error {
	return m.addFn(ctx, s)
}

func (m *mockRepo) GetSubscriptions(ctx context.Context, q models.CalculateQuery) ([]int, error) {
	return m.getFn(ctx, q)
}

func TestAddSubscription(t *testing.T) {
	userID := uuid.New()
	price := 1999

	t.Run("success with end date", func(t *testing.T) {
		var captured models.SubscriptionModel
		svc := New(&mockRepo{
			addFn: func(_ context.Context, m models.SubscriptionModel) error {
				captured = m
				return nil
			},
		})

		endDate := "06-2026"
		dto := models.AddSubscriptionDTO{
			Name:      "netflix",
			Price:     &price,
			UserID:    userID,
			StartDate: "01-2026",
			EndDate:   &endDate,
		}

		err := svc.AddSubscription(context.Background(), dto)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if captured.Name != "netflix" {
			t.Errorf("got name %q, want %q", captured.Name, "netflix")
		}
		if captured.Price != 1999 {
			t.Errorf("got price %d, want %d", captured.Price, 1999)
		}
		if captured.UserID != userID {
			t.Errorf("got user_id %v, want %v", captured.UserID, userID)
		}
		if captured.EndDate == nil {
			t.Fatal("expected non-nil end date")
		}
		wantStart := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
		if !captured.StartDate.Equal(wantStart) {
			t.Errorf("start_date got %v, want %v", captured.StartDate, wantStart)
		}
		wantEnd := time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC)
		if !captured.EndDate.Equal(wantEnd) {
			t.Errorf("end_date got %v, want %v", *captured.EndDate, wantEnd)
		}
	})

	t.Run("success without end date", func(t *testing.T) {
		var captured models.SubscriptionModel
		svc := New(&mockRepo{
			addFn: func(_ context.Context, m models.SubscriptionModel) error {
				captured = m
				return nil
			},
		})

		dto := models.AddSubscriptionDTO{
			Name:      "spotify",
			Price:     &price,
			UserID:    userID,
			StartDate: "03-2026",
			EndDate:   nil,
		}

		err := svc.AddSubscription(context.Background(), dto)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if captured.Name != "spotify" {
			t.Errorf("got name %q, want %q", captured.Name, "spotify")
		}
		if captured.EndDate != nil {
			t.Errorf("expected nil end date, got %v", captured.EndDate)
		}
		wantStart := time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC)
		if !captured.StartDate.Equal(wantStart) {
			t.Errorf("start_date got %v, want %v", captured.StartDate, wantStart)
		}
	})

	t.Run("repo error propagates", func(t *testing.T) {
		svc := New(&mockRepo{
			addFn: func(_ context.Context, m models.SubscriptionModel) error {
				return errors.New("db down")
			},
		})

		dto := models.AddSubscriptionDTO{
			Name:      "hbo",
			Price:     &price,
			UserID:    userID,
			StartDate: "01-2026",
		}

		err := svc.AddSubscription(context.Background(), dto)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if err.Error() != "db down" {
			t.Errorf("got error %q, want %q", err.Error(), "db down")
		}
	})

	t.Run("invalid start_date format", func(t *testing.T) {
		svc := New(&mockRepo{
			addFn: func(_ context.Context, m models.SubscriptionModel) error {
				return nil
			},
		})

		dto := models.AddSubscriptionDTO{
			Name:      "hbo",
			Price:     &price,
			UserID:    userID,
			StartDate: "2026-01-15",
		}

		err := svc.AddSubscription(context.Background(), dto)
		if err == nil {
			t.Fatal("expected parsing error, got nil")
		}
	})
}

func TestCalculate(t *testing.T) {
	userID := uuid.New()
	now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)

	t.Run("sums prices correctly", func(t *testing.T) {
		svc := New(&mockRepo{
			getFn: func(_ context.Context, q models.CalculateQuery) ([]int, error) {
				return []int{1000, 2000, 3000}, nil
			},
		})

		total, err := svc.Calculate(context.Background(), models.CalculateQuery{
			Name:      "netflix",
			ClientID:  userID,
			StartDate: &now,
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if total != 6000 {
			t.Errorf("got total %d, want %d", total, 6000)
		}
	})

	t.Run("empty result returns zero", func(t *testing.T) {
		svc := New(&mockRepo{
			getFn: func(_ context.Context, q models.CalculateQuery) ([]int, error) {
				return []int{}, nil
			},
		})

		total, err := svc.Calculate(context.Background(), models.CalculateQuery{
			Name:     "nonexistent",
			ClientID: userID,
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if total != 0 {
			t.Errorf("got total %d, want 0", total)
		}
	})

	t.Run("repo error propagates", func(t *testing.T) {
		svc := New(&mockRepo{
			getFn: func(_ context.Context, q models.CalculateQuery) ([]int, error) {
				return nil, errors.New("query failed")
			},
		})

		_, err := svc.Calculate(context.Background(), models.CalculateQuery{
			Name:     "netflix",
			ClientID: userID,
		})
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if err.Error() != "query failed" {
			t.Errorf("got error %q, want %q", err.Error(), "query failed")
		}
	})
}
