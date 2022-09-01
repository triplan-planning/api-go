package api

import (
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type SpendingPaidFor struct {
	User       primitive.ObjectID `json:"user" bson:"user"`
	ForcePrice uint32             `json:"forcePrice,omitempty" bson:"forcePrice,omitempty"`
	Weight     float32            `json:"weight,omitempty" bson:"weight,omitempty"`
}

type Spending struct {
	Id       primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	Trip     primitive.ObjectID `json:"trip" bson:"trip"`
	PaidBy   primitive.ObjectID `json:"paidBy" bson:"paidBy"`
	PaidFor  []*SpendingPaidFor `json:"paidFor" bson:"paidFor"`
	Amount   uint32             `json:"amount" bson:"amount"`
	Date     time.Time          `json:"date" bson:"date"`
	Category string             `json:"category" bson:"category"`
	Title    string             `json:"title,omitempty" bson:"title,omitempty"`
}

func (s *Spending) Validate() (err error) {
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

func (api *Api) GetTripSpendings(c *fiber.Ctx) error {
	tripId, err := getId(c, c.Params("id"))
	if err != nil {
		return err
	}
	res, err := api.Mongo.Database("triplan").Collection("spendings").Find(c.Context(), bson.M{"trip": tripId}, options.Find().SetSort(bson.M{"_id": -1}))
	if err != nil {
		return err
	}

	var trips []Spending
	err = res.All(c.Context(), &trips)
	if err != nil {
		return err
	}

	return c.JSON(trips)
}

func (api *Api) PostTripSpending(c *fiber.Ctx) error {
	var spending Spending
	err := c.BodyParser(&spending)
	if err != nil {
		return err
	}
	tripId, err := getId(c, c.Params("id"))
	if err != nil {
		return err
	}
	tripRes := api.Mongo.Database("triplan").Collection("trips").FindOne(c.Context(), bson.M{"_id": tripId})
	if tripRes.Err() != nil {
		return c.Status(fiber.StatusBadRequest).JSON(bson.M{"error": "invalid trip id"})
	}

	spending.Trip = tripId
	if err := spending.Validate(); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(bson.M{"error": err.Error()})
	}

	users := map[primitive.ObjectID]bool{spending.PaidBy: true}
	for _, paidFor := range spending.PaidFor {
		users[paidFor.User] = true
	}

	userIds := []primitive.ObjectID{}
	for id := range users {
		userIds = append(userIds, id)
	}

	cnt, err := api.Mongo.Database("triplan").Collection("users").CountDocuments(c.Context(), bson.M{
		"_id": bson.M{"$in": userIds},
	})
	if err != nil {
		return err
	}
	if cnt != int64(len(users)) {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": fmt.Sprintf(`field "users" must be a list of valid users: got %d valid users out of %d`, cnt, len(users)),
		})
	}

	spending.Id = primitive.NilObjectID

	res, err := api.Mongo.Database("triplan").Collection("spendings").InsertOne(c.Context(), spending)
	if err != nil {
		return err
	}
	spending.Id = res.InsertedID.(primitive.ObjectID)

	return c.JSON(spending)
}

func (api *Api) DeleteSpending(c *fiber.Ctx) error {
	tripId, err := getId(c, c.Params("id"))
	if err != nil {
		return err
	}

	_, err = api.Mongo.Database("triplan").Collection("spendings").DeleteOne(c.Context(), bson.M{
		"_id": tripId,
	})

	if err != nil {
		return err
	}

	c.Status(fiber.StatusNoContent)
	return nil
}

func (api *Api) PutSpending(c *fiber.Ctx) error {
	spendingId, err := getId(c, c.Params("id"))
	if err != nil {
		return err
	}

	var spending Spending
	err = c.BodyParser(&spending)
	if err != nil {
		return err
	}
	if err := spending.Validate(); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(bson.M{"error": err.Error()})
	}

	users := []primitive.ObjectID{spending.PaidBy}
	for _, paidFor := range spending.PaidFor {
		users = append(users, paidFor.User)
	}

	cnt, err := api.Mongo.Database("triplan").Collection("users").CountDocuments(c.Context(), bson.M{
		"_id": bson.M{"$in": users},
	})
	if err != nil {
		return err
	}
	if cnt != int64(len(users)) {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": fmt.Sprintf(`field "users" must be a list of valid users: got %d valid users out of %d`, cnt, len(users)),
		})
	}

	spending.Id = spendingId
	if err != nil {
		return err
	}

	res, err := api.Mongo.Database("triplan").Collection("spendings").ReplaceOne(c.Context(), bson.M{"_id": spending.Id}, spending)
	if err != nil {
		return err
	}
	if res.ModifiedCount != 1 {
		return fmt.Errorf("could not update spending")
	}

	return c.JSON(spending)
}
