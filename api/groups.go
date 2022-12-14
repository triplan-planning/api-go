package api

import (
	"fmt"

	"github.com/gofiber/fiber/v2"
	"github.com/triplan-planning/api-go/model"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func (api *Api) GetGroups(c *fiber.Ctx) error {
	limitUser := c.Query("user")
	filter := bson.M{}
	if limitUser != "" {
		uid, err := getId(limitUser)
		if err != nil {
			return err
		}
		filter["users"] = uid
	}
	res, err := api.groupsColl.Find(c.Context(), filter)
	if err != nil {
		return err
	}

	var trips []model.Group
	err = res.All(c.Context(), &trips)
	if err != nil {
		return err
	}

	return c.JSON(trips)
}

func (api *Api) GetGroupInfo(c *fiber.Ctx) error {
	tripId, err := getId(c.Params("id"))
	if err != nil {
		return err
	}

	res := api.groupsColl.FindOne(c.Context(), bson.M{
		"_id": tripId,
	})
	if res.Err() != nil {
		return res.Err()
	}

	var trip model.Group
	err = res.Decode(&trip)
	if err != nil {
		return err
	}

	return c.JSON(trip)
}

func (api *Api) GetUsersFromGroup(c *fiber.Ctx) error {
	tripId, err := getId(c.Params("id"))
	if err != nil {
		return err
	}

	res := api.groupsColl.FindOne(c.Context(), bson.M{
		"_id": tripId,
	})
	if res.Err() != nil {
		return res.Err()
	}
	var trip model.Group
	err = res.Decode(&trip)
	if err != nil {
		return err
	}

	if len(trip.Users) == 0 {
		return c.JSON([]model.User{})
	}
	resUsers, err := api.usersColl.Find(c.Context(), bson.M{
		"_id": bson.M{"$in": trip.Users},
	})
	if err != nil {
		return err
	}
	var users []model.User
	err = resUsers.All(c.Context(), &users)
	if err != nil {
		return err
	}

	return c.JSON(users)
}

func (api *Api) PostGroup(c *fiber.Ctx) error {
	var trip model.Group
	err := c.BodyParser(&trip)
	if err != nil {
		return err
	}
	if trip.Name == "" {
		return fiber.NewError(fiber.StatusBadRequest, `field "name" must be non-empty`)
	}
	if len(trip.Users) == 0 {
		return fiber.NewError(fiber.StatusBadRequest, `field "users" must be non-empty`)
	}
	cnt, err := api.usersColl.CountDocuments(c.Context(), bson.M{
		"_id": bson.M{"$in": trip.Users},
	})
	if err != nil {
		return err
	}
	if cnt != int64(len(trip.Users)) {
		return fmt.Errorf(`%w: field "users" must be a list of valid users: got %d valid users out of %d`, fiber.ErrBadRequest, cnt, len(trip.Users))
	}

	trip.Id = primitive.NilObjectID

	res, err := api.groupsColl.InsertOne(c.Context(), trip)
	if err != nil {
		return err
	}
	trip.Id = res.InsertedID.(primitive.ObjectID)

	return c.JSON(trip)
}

func (api *Api) DeleteGroup(c *fiber.Ctx) error {
	tripId, err := getId(c.Params("id"))
	if err != nil {
		return err
	}

	_, err = api.groupsColl.DeleteOne(c.Context(), bson.M{
		"_id": tripId,
	})

	if err != nil {
		return err
	}

	c.Status(fiber.StatusNoContent)
	return nil
}

func (api *Api) PutGroup(c *fiber.Ctx) error {
	tripId, err := getId(c.Params("id"))
	if err != nil {
		return err
	}

	var trip model.Group
	err = c.BodyParser(&trip)
	if err != nil {
		return err
	}
	if trip.Name == "" {
		return fiber.NewError(fiber.StatusBadRequest, `field "name" must be non-empty`)
	}
	if len(trip.Users) == 0 {
		return fiber.NewError(fiber.StatusBadRequest, `field "users" must be non-empty`)
	}
	cnt, err := api.usersColl.CountDocuments(c.Context(), bson.M{
		"_id": bson.M{"$in": trip.Users},
	})
	if err != nil {
		return err
	}
	if cnt != int64(len(trip.Users)) {
		return fmt.Errorf(`%w: field "users" must be a list of valid users: got %d valid users out of %d`, fiber.ErrBadRequest, cnt, len(trip.Users))
	}

	trip.Id = tripId
	if err != nil {
		return err
	}

	res, err := api.groupsColl.ReplaceOne(c.Context(), bson.M{
		"_id": trip.Id,
	}, trip)
	if err != nil {
		return err
	}
	if res.ModifiedCount != 1 {
		return fmt.Errorf("could not update trip")
	}

	return c.JSON(trip)
}
