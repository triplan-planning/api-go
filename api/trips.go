package api

import (
	"fmt"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Trip struct {
	Id          primitive.ObjectID   `json:"id" bson:"_id,omitempty"`
	Name        string               `json:"name,omitempty" bson:"name,omitempty"`
	Description string               `json:"description,omitempty" bson:"description,omitempty"`
	Users       []primitive.ObjectID `json:"users,omitempty" bson:"users,omitempty"`
}

func (api *Api) GetTrips(c *fiber.Ctx) error {
	limitUser := c.Query("user")
	filter := bson.M{}
	if limitUser != "" {
		uid, err := getId(c, limitUser)
		if err != nil {
			return err
		}
		filter["users"] = uid
	}
	res, err := api.Mongo.Database("triplan").
		Collection("trips").Find(c.Context(), filter)
	if err != nil {
		return err
	}

	var trips []Trip
	err = res.All(c.Context(), &trips)
	if err != nil {
		return err
	}

	return c.JSON(trips)
}

func (api *Api) PostTrip(c *fiber.Ctx) error {
	var trip Trip
	err := c.BodyParser(&trip)
	if err != nil {
		return err
	}
	if trip.Name == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": `field "name" must be non-empty`,
		})
	}
	if len(trip.Users) == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": `field "users" must be non-empty`,
		})
	}
	cnt, err := api.Mongo.Database("triplan").Collection("users").CountDocuments(c.Context(), bson.M{
		"_id": bson.M{"$in": trip.Users},
	})
	if err != nil {
		return err
	}
	if cnt != int64(len(trip.Users)) {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": fmt.Sprintf(`field "users" must be a list of valid users: got %d valid users out of %d`, cnt, len(trip.Users)),
		})
	}

	trip.Id = primitive.NilObjectID

	res, err := api.Mongo.Database("triplan").Collection("trips").InsertOne(c.Context(), trip)
	if err != nil {
		return err
	}
	trip.Id = res.InsertedID.(primitive.ObjectID)

	return c.JSON(trip)
}

func (api *Api) DeleteTrip(c *fiber.Ctx) error {
	tripId, err := getId(c, c.Params("id"))
	if err != nil {
		return err
	}

	_, err = api.Mongo.Database("triplan").
		Collection("trips").DeleteOne(c.Context(), bson.M{
		"_id": tripId,
	})

	if err != nil {
		return err
	}

	c.Status(fiber.StatusNoContent)
	return nil
}

func (api *Api) PutTrip(c *fiber.Ctx) error {
	tripId, err := getId(c, c.Params("id"))
	if err != nil {
		return err
	}

	var trip Trip
	err = c.BodyParser(&trip)
	if err != nil {
		return err
	}
	if trip.Name == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": `field "name" must be non-empty`,
		})
	}
	if len(trip.Users) == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": `field "users" must be non-empty`,
		})
	}
	cnt, err := api.Mongo.Database("triplan").Collection("users").CountDocuments(c.Context(), bson.M{
		"_id": bson.M{"$in": trip.Users},
	})
	if err != nil {
		return err
	}
	if cnt != int64(len(trip.Users)) {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": fmt.Sprintf(`field "users" must be a list of valid users: got %d valid users out of %d`, cnt, len(trip.Users)),
		})
	}

	trip.Id = tripId
	if err != nil {
		return err
	}

	res, err := api.Mongo.Database("triplan").
		Collection("trips").ReplaceOne(c.Context(), bson.M{
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
