package api

import (
	"fmt"

	"github.com/gofiber/fiber/v2"
	"github.com/triplan-planning/api-go/model"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// @Summary      Returns all the spending from this trip
// @Accept       json
// @Param        id   path      string  true  "Group ID"
// @Success      200  {object}  model.Transaction
// @Router       /groups/{id}/transactions [get]
func (api *Api) GetGroupTransactions(c *fiber.Ctx) error {
	tripId, err := getId(c.Params("id"))
	if err != nil {
		return err
	}
	res, err := api.transactionsColl.Find(c.Context(), model.Transaction{Group: tripId}, options.Find().SetSort(bson.M{"_id": -1}))
	if err != nil {
		return err
	}

	var transactions []model.Transaction
	err = res.All(c.Context(), &transactions)
	if err != nil {
		return err
	}

	return c.JSON(transactions)
}

// @Summary      Creates a transaction
// @Accept       json
// @Param        id   path      string  true  "Group ID"
// @Param        transaction  body      model.Transaction  true  "The transaction to create"
// @Success      200  {object}  model.Transaction
// @Router       /groups/{id}/transactions [post]
func (api *Api) PostGroupTransaction(c *fiber.Ctx) error {
	var transaction model.Transaction
	err := c.BodyParser(&transaction)
	if err != nil {
		return err
	}
	groupId, err := getId(c.Params("id"))
	if err != nil {
		return err
	}
	groupRes := api.groupsColl.FindOne(c.Context(), bson.M{"_id": groupId})
	if groupRes.Err() != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid group id")
	}

	transaction.Group = groupId
	if err := transaction.Validate(); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}

	users := transaction.Users()

	filter := bson.M{
		"_id":   groupId,
		"users": bson.M{"$in": users},
	}
	cnt, err := api.groupsColl.CountDocuments(c.Context(), filter)
	if err != nil {
		return err
	}
	if cnt != int64(len(users)) {
		return fmt.Errorf(` %w: field "users" must be a list of valid group members: got %d valid users out of %d`, fiber.ErrBadRequest, cnt, len(users))
	}

	transaction.Id = primitive.NilObjectID
	err = transaction.ComputePrices()
	if err != nil {
		return err
	}

	res, err := api.transactionsColl.InsertOne(c.Context(), transaction)
	if err != nil {
		return err
	}
	transaction.Id = res.InsertedID.(primitive.ObjectID)

	return c.JSON(transaction)
}

// @Summary      Deletes a transaction
// @Param        id   path      string  true  "Transaction ID"
// @Success      204
// @Router       /transactions/{id} [delete]
func (api *Api) DeleteTransaction(c *fiber.Ctx) error {
	tripId, err := getId(c.Params("id"))
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

// @Summary      Updates a transaction
// @Accept       json
// @Param        id   path      string  true  "Transaction ID"
// @Param        transaction  body      model.Transaction  true  "The transaction to update"
// @Success      200  {object}  model.Transaction
// @Router       /transactions/{id} [put]
func (api *Api) PutTransaction(c *fiber.Ctx) error {
	spendingId, err := getId(c.Params("id"))
	if err != nil {
		return err
	}

	var transaction model.Transaction
	err = c.BodyParser(&transaction)
	if err != nil {
		return err
	}
	if err := transaction.Validate(); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}

	users := transaction.Users()

	cnt, err := api.usersColl.CountDocuments(c.Context(), bson.M{
		"_id": bson.M{"$in": users},
	})
	if err != nil {
		return err
	}
	if cnt != int64(len(users)) {
		return fiber.NewError(fiber.StatusBadRequest, fmt.Sprintf(`field "users" must be a list of valid users: got %d valid users out of %d`, cnt, len(users)))
	}

	transaction.Id = spendingId
	if err != nil {
		return err
	}

	err = transaction.ComputePrices()
	if err != nil {
		return err
	}

	res, err := api.transactionsColl.ReplaceOne(c.Context(), bson.M{"_id": transaction.Id}, transaction)
	if err != nil {
		return err
	}
	if res.ModifiedCount != 1 {
		return fmt.Errorf("could not update spending")
	}

	return c.JSON(transaction)
}
