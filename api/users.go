package api

import (
	"fmt"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type User struct {
	Id   primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	Name string             `json:"name" bson:"name,omitempty"`
}

func (api *Api) GetUsers(c *fiber.Ctx) error {
	res, err := api.Mongo.Database("triplan").
		Collection("users").Find(c.Context(), bson.M{})
	if err != nil {
		return err
	}

	var users []User
	err = res.All(c.Context(), &users)
	if err != nil {
		return err
	}

	return c.JSON(users)
}

func (api *Api) PostUser(c *fiber.Ctx) error {
	var user User
	err := c.BodyParser(&user)
	if err != nil {
		return err
	}
	if user.Name == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": `field "name" must be non-empty`,
		})
	}
	user.Id = primitive.NilObjectID

	res, err := api.Mongo.Database("triplan").
		Collection("users").InsertOne(c.Context(), user)
	if err != nil {
		return err
	}
	user.Id = res.InsertedID.(primitive.ObjectID)

	return c.JSON(user)
}

func (api *Api) DeleteUser(c *fiber.Ctx) error {
	userId, err := getId(c, c.Params("id"))
	if err != nil {
		return err
	}

	_, err = api.Mongo.Database("triplan").
		Collection("users").DeleteOne(c.Context(), bson.M{
		"_id": userId,
	})

	if err != nil {
		return err
	}

	c.Status(fiber.StatusNoContent)
	return nil
}

func (api *Api) PutUser(c *fiber.Ctx) error {
	userId, err := getId(c, c.Params("id"))
	if err != nil {
		return err
	}

	var user User
	err = c.BodyParser(&user)
	if err != nil {
		return err
	}
	if user.Name == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": `field "name" must be non-empty`,
		})
	}
	user.Id = userId
	if err != nil {
		return err
	}

	res, err := api.Mongo.Database("triplan").
		Collection("users").ReplaceOne(c.Context(), bson.M{
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