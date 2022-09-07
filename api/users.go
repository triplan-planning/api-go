package api

import (
	"fmt"

	"github.com/gofiber/fiber/v2"
	"github.com/triplan-planning/api-go/model"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func (api *Api) GetUsers(c *fiber.Ctx) error {
	res, err := api.usersColl.Find(c.Context(), bson.M{})
	if err != nil {
		return err
	}

	var users []model.User
	err = res.All(c.Context(), &users)
	if err != nil {
		return err
	}

	return c.JSON(users)
}

func (api *Api) GetUserInfo(c *fiber.Ctx) error {
	userId, err := getId(c.Params("id"))
	if err != nil {
		return err
	}

	res := api.usersColl.FindOne(c.Context(), bson.M{
		"_id": userId,
	})
	if res.Err() != nil {
		return res.Err()
	}

	var user model.User
	err = res.Decode(&user)
	if err != nil {
		return err
	}

	return c.JSON(user)
}

func (api *Api) PostUser(c *fiber.Ctx) error {
	var user model.User
	err := c.BodyParser(&user)
	if err != nil {
		return err
	}
	if user.Name == "" {
		return fiber.NewError(fiber.StatusBadRequest, `field "name" must be non-empty`)
	}
	user.Id = primitive.NilObjectID

	res, err := api.usersColl.InsertOne(c.Context(), user)
	if err != nil {
		return err
	}
	user.Id = res.InsertedID.(primitive.ObjectID)

	return c.JSON(user)
}

func (api *Api) DeleteUser(c *fiber.Ctx) error {
	userId, err := getId(c.Params("id"))
	if err != nil {
		return err
	}

	_, err = api.usersColl.DeleteOne(c.Context(), bson.M{
		"_id": userId,
	})

	if err != nil {
		return err
	}

	c.Status(fiber.StatusNoContent)
	return nil
}

func (api *Api) PutUser(c *fiber.Ctx) error {
	userId, err := getId(c.Params("id"))
	if err != nil {
		return err
	}

	var user model.User
	err = c.BodyParser(&user)
	if err != nil {
		return err
	}
	if user.Name == "" {
		return fiber.NewError(fiber.StatusBadRequest, `field "name" must be non-empty`)
	}
	user.Id = userId
	if err != nil {
		return err
	}

	res, err := api.usersColl.ReplaceOne(c.Context(), bson.M{
		"_id": user.Id,
	}, user)
	if err != nil {
		return err
	}
	if res.ModifiedCount != 1 {
		return fmt.Errorf("could not update user")
	}

	return c.JSON(user)
}
