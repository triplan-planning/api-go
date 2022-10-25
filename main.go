package main

import (
	"context"
	"errors"
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
	routes := api.New(db)

	app := fiber.New(fiber.Config{
		ErrorHandler: func(ctx *fiber.Ctx, err error) error {
			code := fiber.StatusInternalServerError
			var e *fiber.Error
			if errors.As(err, &e) {
				code = e.Code
			}
			ctx.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSONCharsetUTF8)
			return ctx.Status(code).JSON(fiber.Map{"error": err.Error()})
		},
	})
	app.Use(cors.New())

	app.Get("/", routes.HomeStats)
	app.Get("/doc/*", swagger.HandlerDefault)
	users := app.Group("/users")
	users.Get("", routes.GetUsers)
	users.Get("/:id", routes.GetUserInfo)
	users.Post("", routes.PostUser)
	users.Delete("/:id", routes.DeleteUser)
	users.Put("/:id", routes.PutUser)

	groups := app.Group("/groups")
	groups.Get("", routes.GetGroups)
	groups.Get("/:id/users", routes.GetUsersFromGroup)
	groups.Get("/:id", routes.GetGroupInfo)
	groups.Post("", routes.PostGroup)
	groups.Delete("/:id", routes.DeleteGroup)
	groups.Put("/:id", routes.PutGroup)

	groups.Get("/:id/transactions", routes.GetGroupTransactions)
	groups.Post("/:id/transactions", routes.PostGroupTransaction)
	transactions := app.Group("/transactions")
	transactions.Delete("/:id", routes.DeleteTransaction)
	transactions.Put("/:id", routes.PutTransaction)

	app.Listen("0.0.0.0" + getPort())
}
