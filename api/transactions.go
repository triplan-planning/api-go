package api

import (
	"fmt"

	"github.com/gofiber/fiber/v2"
	"github.com/triplan-planning/api-go/model"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// GetTripTransactions returns the spendings from a trip
// @Summary      Returns all the spending from this trip
// @Accept       json
// @Param        id   path      string  true  "Trip ID"
// @Success      200  {object}  Transaction
// @Router       /trips/{id}/spendings [get]
func (api *Api) GetTripTransactions(c *fiber.Ctx) error {
	tripId, err := getId(c, c.Params("id"))
	if err != nil {
		return err
	}
	res, err := api.transactionsColl.Find(c.Context(), bson.M{"trip": tripId}, options.Find().SetSort(bson.M{"_id": -1}))
	if err != nil {
		return err
	}

	var trips []model.Transaction
	err = res.All(c.Context(), &trips)
	if err != nil {
		return err
	}

	return c.JSON(trips)
}

// PostTripTransaction creates a spending
// @Summary      Creates a spending
// @Accept       json
// @Param        id   path      string  true  "Trip ID"
// @Param        spending  body      Transaction  true  "The spending to create"
// @Success      200  {object}  Transaction
// @Router       /trips/{id}/spendings [post]
func (api *Api) PostTripTransaction(c *fiber.Ctx) error {
	var spending model.Transaction
	err := c.BodyParser(&spending)
	if err != nil {
		return err
	}
	tripId, err := getId(c, c.Params("id"))
	if err != nil {
		return err
	}
	tripRes := api.groupsColl.FindOne(c.Context(), bson.M{"_id": tripId})
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

	cnt, err := api.usersColl.CountDocuments(c.Context(), bson.M{
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

	res, err := api.transactionsColl.InsertOne(c.Context(), spending)
	if err != nil {
		return err
	}
	spending.Id = res.InsertedID.(primitive.ObjectID)

	return c.JSON(spending)
}

func (api *Api) DeleteTransaction(c *fiber.Ctx) error {
	tripId, err := getId(c, c.Params("id"))
	if err != nil {
		return err
	}

	_, err = api.transactionsColl.DeleteOne(c.Context(), bson.M{
		"_id": tripId,
	})

	if err != nil {
		return err
	}

	c.Status(fiber.StatusNoContent)
	return nil
}

func (api *Api) PutTransaction(c *fiber.Ctx) error {
	spendingId, err := getId(c, c.Params("id"))
	if err != nil {
		return err
	}

	var spending model.Transaction
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

	cnt, err := api.usersColl.CountDocuments(c.Context(), bson.M{
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

	res, err := api.transactionsColl.ReplaceOne(c.Context(), bson.M{"_id": spending.Id}, spending)
	if err != nil {
		return err
	}
	if res.ModifiedCount != 1 {
		return fmt.Errorf("could not update spending")
	}

	return c.JSON(spending)
}
