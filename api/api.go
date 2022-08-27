package api

import (
	"errors"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type Api struct {
	Mongo *mongo.Client
}

func getId(c *fiber.Ctx, idstring string) (primitive.ObjectID, error) {
	userId, err := primitive.ObjectIDFromHex(idstring)
	if err != nil {
		if errors.Is(err, primitive.ErrInvalidHex) {
			return primitive.NilObjectID, c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": `id must be a valid id`,
			})
		}
		return primitive.NilObjectID, err
	}
	return userId, err
}
