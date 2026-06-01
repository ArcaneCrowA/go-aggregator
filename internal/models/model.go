package models

import (
	"time"

	"github.com/google/uuid"
)

type SubscriptionModel struct {
	Name      string     `db:"name"`
	Price     int        `db:"price"`
	UserID    uuid.UUID  `db:"user_id"`
	StartDate time.Time  `db:"start_date"`
	EndDate   *time.Time `db:"end_date"`
}
