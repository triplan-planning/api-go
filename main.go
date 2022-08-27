package main

import (
	"context"
	"fiber/api"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"

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

	app.Listen("0.0.0.0" + getPort())
}
