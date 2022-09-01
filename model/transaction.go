package model

import (
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type TransactionTarget struct {
	User       primitive.ObjectID `json:"user" bson:"user"`
	ForcePrice uint32             `json:"forcePrice,omitempty" bson:"forcePrice,omitempty"`
	Weight     uint32             `json:"weight,omitempty" bson:"weight,omitempty"`
}

type Transaction struct {
	Id       primitive.ObjectID   `json:"id" bson:"_id,omitempty"`
	Trip     primitive.ObjectID   `json:"trip" bson:"trip"`
	PaidBy   primitive.ObjectID   `json:"paidBy" bson:"paidBy"`
	PaidFor  []*TransactionTarget `json:"paidFor" bson:"paidFor"`
	Amount   uint32               `json:"amount" bson:"amount"`
	Date     time.Time            `json:"date" bson:"date"`
	Category string               `json:"category" bson:"category"`
	Title    string               `json:"title,omitempty" bson:"title,omitempty"`
}

func (s *Transaction) Validate() (err error) {
	if s.Amount == 0 {
		return fmt.Errorf(`field "amount" must be non-zero`)
	}
	if s.Trip.IsZero() {
		return fmt.Errorf(`field "trip" must be filled`)
	}
	if s.PaidBy.IsZero() {
		return fmt.Errorf(`field "paidBy" must be filled`)
	}
	if len(s.PaidFor) == 0 {
		return fmt.Errorf(`field "paidFor" must have some values`)
	}
	if s.Date.IsZero() {
		return fmt.Errorf(`field "date" must have be filled`)
	}
	if s.Category == "" {
		return fmt.Errorf(`field "category" must have be filled`)
	}

	return nil
}
