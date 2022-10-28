package model

import (
	"errors"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type TransactionTarget struct {
	User          primitive.ObjectID `json:"user" bson:"user"`
	ForcePrice    uint32             `json:"forcePrice,omitempty" bson:"forcePrice,omitempty"`
	Weight        uint32             `json:"weight,omitempty" bson:"weight,omitempty"`
	ComputedPrice uint32             `json:"computedPrice,omitempty" bson:"computedPrice,omitempty"`
}

type Transaction struct {
	Id       primitive.ObjectID   `json:"id" bson:"_id,omitempty"`
	Group    primitive.ObjectID   `json:"group" bson:"group,omitempty"`
	PaidBy   primitive.ObjectID   `json:"paidBy" bson:"paidBy,omitempty"`
	PaidFor  []*TransactionTarget `json:"paidFor" bson:"paidFor,omitempty"`
	Amount   uint32               `json:"amount" bson:"amount,omitempty"`
	Date     time.Time            `json:"date" bson:"date,omitempty"`
	Category string               `json:"category" bson:"category,omitempty"`
	Title    string               `json:"title,omitempty" bson:"title,omitempty"`
}

func (s *Transaction) Validate() (err error) {
	if s.Amount == 0 {
		return fmt.Errorf(`field "amount" must be non-zero`)
	}
	if s.Group.IsZero() {
		return fmt.Errorf(`field "group" must be filled`)
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

func (s *Transaction) Users() []primitive.ObjectID {
	users := map[primitive.ObjectID]bool{s.PaidBy: true}
	for _, paidFor := range s.PaidFor {
		users[paidFor.User] = true
	}

	userIds := []primitive.ObjectID{}
	for id := range users {
		userIds = append(userIds, id)
	}

	return userIds
}

func (s *Transaction) ComputePrices() (err error) {
	rest := s.Amount
	totalWeights := uint32(0)
	for _, t := range s.PaidFor {
		if t.ForcePrice > rest {
			return errors.New("force prices are higher than the transaction amount, please fix")
		}
		rest -= t.ForcePrice
		totalWeights += t.Weight
	}
	if rest == 0 {
		return nil
	}

	toSplit := rest

	for _, t := range s.PaidFor {
		t.ComputedPrice = 0
		if t.ForcePrice != 0 {
			t.ComputedPrice = t.ForcePrice
		}
		if t.Weight != 0 {
			part := toSplit * t.Weight / totalWeights
			t.ComputedPrice += part
			rest -= part
		}
	}

	// if part of the amount could not be parted between members, increment each person computed price by the smallest unit
	if rest > 0 {
		for _, t := range s.PaidFor {
			t.ComputedPrice += 1
		}

		roundValue := uint32(len(s.PaidFor))
		// prevent from substracting rest lower than zero
		if rest > roundValue {
			rest -= roundValue
		} else {
			rest = 0
		}
	}

	if rest > 0 {
		return errors.New("computations are fucked up")
	}

	return nil
}
