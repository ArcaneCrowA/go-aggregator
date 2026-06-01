package models

import (
	"testing"

	"github.com/google/uuid"
)

func TestAddSubscriptionDTO_Validate(t *testing.T) {
	validUUID := uuid.New()
	price := 999

	t.Run("valid DTO passes", func(t *testing.T) {
		dto := AddSubscriptionDTO{
			Name:      "netflix",
			Price:     &price,
			UserID:    validUUID,
			StartDate: "01-2026",
		}
		if err := dto.Validate(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("valid DTO with optional end_date passes", func(t *testing.T) {
		endDate := "06-2026"
		dto := AddSubscriptionDTO{
			Name:      "spotify",
			Price:     &price,
			UserID:    validUUID,
			StartDate: "01-2026",
			EndDate:   &endDate,
		}
		if err := dto.Validate(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("missing name fails", func(t *testing.T) {
		dto := AddSubscriptionDTO{
			Price:     &price,
			UserID:    validUUID,
			StartDate: "01-2026",
		}
		if err := dto.Validate(); err == nil {
			t.Fatal("expected error, got nil")
		}
	})

	t.Run("missing price fails", func(t *testing.T) {
		dto := AddSubscriptionDTO{
			Name:      "netflix",
			UserID:    validUUID,
			StartDate: "01-2026",
		}
		if err := dto.Validate(); err == nil {
			t.Fatal("expected error, got nil")
		}
	})

	t.Run("missing user_id fails", func(t *testing.T) {
		dto := AddSubscriptionDTO{
			Name:      "netflix",
			Price:     &price,
			StartDate: "01-2026",
		}
		if err := dto.Validate(); err == nil {
			t.Fatal("expected error, got nil")
		}
	})

	t.Run("invalid uuid fails", func(t *testing.T) {
		dto := AddSubscriptionDTO{
			Name:      "netflix",
			Price:     &price,
			UserID:    uuid.UUID{},
			StartDate: "01-2026",
		}
		if err := dto.Validate(); err == nil {
			t.Fatal("expected error, got nil")
		}
	})

	t.Run("missing start_date fails", func(t *testing.T) {
		dto := AddSubscriptionDTO{
			Name:   "netflix",
			Price:  &price,
			UserID: validUUID,
		}
		if err := dto.Validate(); err == nil {
			t.Fatal("expected error, got nil")
		}
	})

	t.Run("bad date format fails", func(t *testing.T) {
		dto := AddSubscriptionDTO{
			Name:      "netflix",
			Price:     &price,
			UserID:    validUUID,
			StartDate: "2026-01-15",
		}
		if err := dto.Validate(); err == nil {
			t.Fatal("expected error, got nil")
		}
	})
}
