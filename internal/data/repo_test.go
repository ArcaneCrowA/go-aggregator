package data

import (
	"strings"
	"testing"
	"time"

	"github.com/ArcaneCrowA/go-aggregator/internal/models"
	"github.com/google/uuid"
	"github.com/huandu/go-sqlbuilder"
)

func TestAddSubscriptionSQL(t *testing.T) {
	t.Run("uses postgres placeholders with end_date", func(t *testing.T) {
		now := time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC)
		end := time.Date(2027, 1, 1, 0, 0, 0, 0, time.UTC)
		m := models.SubscriptionModel{
			UserID:    uuid.New(),
			Name:      "yandex",
			Price:     400,
			StartDate: now,
			EndDate:   &end,
		}

		ib := sqlbuilder.PostgreSQL.NewInsertBuilder()
		ib.InsertInto("subscriptions")
		ib.Cols("user_id", "name", "price", "start_date", "end_date")
		ib.Values(m.UserID, m.Name, m.Price, m.StartDate, m.EndDate)
		q, args := ib.Build()

		if len(args) != 5 {
			t.Fatalf("expected 5 args, got %d", len(args))
		}
		if args[0] != m.UserID {
			t.Errorf("arg[0] = %v, want %v", args[0], m.UserID)
		}
		if args[1] != "yandex" {
			t.Errorf("arg[1] = %v, want yandex", args[1])
		}
		if args[2] != 400 {
			t.Errorf("arg[2] = %v, want 400", args[2])
		}
		if !strings.Contains(q, "$1") || !strings.Contains(q, "$5") {
			t.Errorf("expected $1..$5 placeholders in query, got: %s", q)
		}
	})

	t.Run("handles nil end_date gracefully", func(t *testing.T) {
		now := time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC)
		m := models.SubscriptionModel{
			UserID:    uuid.New(),
			Name:      "spotify",
			Price:     999,
			StartDate: now,
			EndDate:   nil,
		}

		ib := sqlbuilder.PostgreSQL.NewInsertBuilder()
		ib.InsertInto("subscriptions")
		ib.Cols("user_id", "name", "price", "start_date", "end_date")
		ib.Values(m.UserID, m.Name, m.Price, m.StartDate, m.EndDate)
		q, args := ib.Build()

		if len(args) != 5 {
			t.Fatalf("expected 5 args, got %d", len(args))
		}
		if !strings.Contains(q, "$5") {
			t.Errorf("expected $5 placeholder for end_date in query, got: %s", q)
		}
	})
}

func TestGetSubscriptionsSQL(t *testing.T) {
	t.Run("uses postgres placeholders with dollar signs", func(t *testing.T) {
		sel := sqlbuilder.PostgreSQL.NewSelectBuilder()
		sel.Select("price")
		sel.From("subscriptions")
		sel.Where(
			sel.Equal("user_id", uuid.New()),
			sel.Like("name", "netflix"),
		)

		q, args := sel.Build()

		if len(args) != 2 {
			t.Fatalf("expected 2 args, got %d", len(args))
		}
		if !strings.Contains(q, "$1") || !strings.Contains(q, "$2") {
			t.Errorf("expected $1 and $2 placeholders in query, got: %s", q)
		}
	})
}
