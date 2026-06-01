package models

import (
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
)

var validate = validator.New()

type AddSubscriptionDTO struct {
	Name      string    `json:"service_name" validate:"required"`
	Price     *int      `json:"price" validate:"required"`
	UserID    uuid.UUID `json:"user_id" validate:"required,uuid"`
	StartDate string    `json:"start_date" validate:"required,datetime=01-2006"`
	EndDate   *string   `json:"end_date" validate:"omitempty,datetime=01-2006"`
}

func (a AddSubscriptionDTO) Validate() error {
	return validate.Struct(a)
}

type CalculateQuery struct {
	Name      string
	ClientID  uuid.UUID
	StartDate *time.Time
	EndDate   *time.Time
}
