package api

import (
	"errors"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func New(db *mongo.Client) *Api {
	return &Api{
		Mongo:            db,
		groupsColl:       db.Database("triplan").Collection("groups"),
		usersColl:        db.Database("triplan").Collection("users"),
		transactionsColl: db.Database("triplan").Collection("transactions"),
	}
}

type Api struct {
	Mongo            *mongo.Client
	groupsColl       *mongo.Collection
	usersColl        *mongo.Collection
	transactionsColl *mongo.Collection
}

func getId(idstring string) (primitive.ObjectID, error) {
	userId, err := primitive.ObjectIDFromHex(idstring)
	if err != nil {
		if errors.Is(err, primitive.ErrInvalidHex) {
			return primitive.NilObjectID, fiber.NewError(fiber.StatusBadRequest, "id must be a valid id")
		}
		return primitive.NilObjectID, err
	}
	return userId, err
}

func (api *Api) HomeStats(c *fiber.Ctx) error {
	res := api.Mongo.Database("stats").Collection("http_calls").FindOneAndUpdate(c.Context(), bson.M{"_id": "/"}, bson.M{"$inc": bson.M{"count": 1}}, options.FindOneAndUpdate().SetUpsert(true).SetReturnDocument(options.After))

	if res.Err() != nil {
		return fiber.NewError(fiber.StatusBadRequest, "😢 could not do it: "+res.Err().Error())
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
