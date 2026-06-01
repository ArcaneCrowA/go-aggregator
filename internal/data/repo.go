package data

import (
	"context"
	"errors"
	"time"

	"github.com/ArcaneCrowA/go-aggregator/internal/models"
	"github.com/huandu/go-sqlbuilder"
	"github.com/jackc/pgx/v5/pgxpool"
)

type repo struct {
	db *pgxpool.Pool
}

func New(db *pgxpool.Pool) *repo {
	return &repo{
		db: db,
	}
}

func (r *repo) AddSubscription(ctx context.Context, m models.SubscriptionModel) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	ib := sqlbuilder.PostgreSQL.NewInsertBuilder()
	ib.InsertInto("subscriptions")
	ib.Cols("user_id", "name", "price", "start_date", "end_date")
	ib.Values(m.UserID, m.Name, m.Price, m.StartDate, m.EndDate)
	query, args := ib.Build()

	res, err := r.db.Exec(ctx, query, args...)
	if err != nil {
		return err
	}
	if res.RowsAffected() != 1 {
		return errors.New("didn't affect table")
	}

	return nil
}

func (r *repo) GetSubscriptions(ctx context.Context, m models.CalculateQuery) ([]int, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	sel := sqlbuilder.PostgreSQL.NewSelectBuilder()
	sel.Select("price")
	sel.From("subscriptions")

	sel.Where(sel.Equal("user_id", m.ClientID))
	if m.Name != "" {
		sel.Where(sel.Like("name", m.Name))
	}
	if m.StartDate != nil {
		sel.Where(sel.GreaterEqualThan("start_date", *m.StartDate))
	}
	if m.EndDate != nil {
		sel.Where(sel.Or(
			sel.IsNull("end_date"),
			sel.LessEqualThan("end_date", *m.EndDate),
		))
	}

	query, args := sel.Build()

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ms []int
	for rows.Next() {
		var price int
		err := rows.Scan(&price)
		if err != nil {
			return nil, err
		}
		ms = append(ms, price)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return ms, nil
}
