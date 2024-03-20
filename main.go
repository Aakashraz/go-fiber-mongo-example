package main

import (
	"context"
	"fmt"
	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
	"time"
)

type MongoInstance struct {
	Client *mongo.Client
	Db     *mongo.Database
}

var mg MongoInstance

const dbName = "fiber-hrms"
const mongoURI = "mongodb://localhost:27017" + dbName

type Employee struct {
	ID     string  `json:"id,omitempty" bson:"_id, omitempty"`
	Name   string  `json:"name"`
	Salary float64 `json:"salary"`
	Age    float64 `json:"age"`
}

func connect() error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(mongoURI))
	if err != nil {
		log.Printf("error while connecting to mongoDB: %s", err)
		return err
	}

	db := client.Database(dbName)

	mg = MongoInstance{
		Client: client,
		Db:     db,
	}
	return nil
}

func main() {
	fmt.Println("HRM with go FIBER...........")

	if err := connect(); err != nil {
		log.Fatal(err)
	}
	app := fiber.New()
	app.Get("/employee", func(ctx *fiber.Ctx) error {

	})
	app.Post("/employee")
	app.Put("/employee/:id")
	app.Delete("/employee/:id")
}
