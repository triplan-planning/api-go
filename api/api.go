package api

import (
	"errors"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
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

func (api *Api) HomeStats(c *fiber.Ctx) error {
	res := api.Mongo.Database("stats").Collection("http_calls").FindOneAndUpdate(c.Context(), bson.M{"_id": "/"}, bson.M{"$inc": bson.M{"count": 1}}, options.FindOneAndUpdate().SetUpsert(true).SetReturnDocument(options.After))

	if res.Err() != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "😢 could not do it: " + res.Err().Error(),
		})
	}
	out := map[string]any{}
	err := res.Decode(&out)
	if err != nil {
		return err
	}
	return c.JSON(fiber.Map{
		"message": "Wow, triplan !",
		"calls":   out["count"],
	})
}
