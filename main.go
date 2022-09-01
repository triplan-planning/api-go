package main

import (
	"context"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"

	"github.com/gofiber/swagger"
	"github.com/triplan-planning/api-go/api"
	_ "github.com/triplan-planning/api-go/docs"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

func getPort() string {
	port := os.Getenv("PORT")
	if port == "" {
		port = ":3000"
	} else {
		port = ":" + port
	}

	return port
}

func getMongo() *mongo.Client {
	mongourl, ok := os.LookupEnv("MONGO_URL")
	if !ok {
		panic("env variable MONGO_URL must be present")
	}
	// Create a new client and connect to the server
	client, err := mongo.Connect(context.TODO(), options.Client().ApplyURI(mongourl))
	if err != nil {
		panic(err)
	}
	// Ping the primary
	if err := client.Ping(context.TODO(), readpref.Primary()); err != nil {
		panic(err)
	}

	return client
}

// @title           Triplan API
// @version         1.0
// @description     Triplan API POC
// @license.name	Unlicense
func main() {
	db := getMongo()
	defer func() {
		if err := db.Disconnect(context.TODO()); err != nil {
			panic(err)
		}
	}()

	app := fiber.New()
	app.Use(cors.New())
	routes := api.Api{
		Mongo: db,
	}
	app.Get("/", routes.HomeStats)
	app.Get("/doc/*", swagger.HandlerDefault)
	users := app.Group("/users")
	users.Get("", routes.GetUsers)
	users.Get("/:id", routes.GetUserInfo)
	users.Post("", routes.PostUser)
	users.Delete("/:id", routes.DeleteUser)
	users.Put("/:id", routes.PutUser)

	trips := app.Group("/trips")
	trips.Get("", routes.GetTrips)
	trips.Get("/:id", routes.GetTripInfo)
	trips.Post("", routes.PostTrip)
	trips.Delete("/:id", routes.DeleteTrip)
	trips.Put("/:id", routes.PutTrip)

	trips.Get("/:id/spendings", routes.GetTripSpendings)
	trips.Post("/:id/spendings", routes.PostTripSpending)
	spendings := app.Group("/spendings")
	spendings.Delete("/:id", routes.DeleteSpending)
	spendings.Put("/:id", routes.PutSpending)

	app.Listen("0.0.0.0" + getPort())
}
