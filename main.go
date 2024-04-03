package main

import (
	"context"
	"fmt"
	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson"
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
	//client: This variable holds a pointer to a mongo.Client object, which represents the connection to the MongoDB server.
	//You can use this client object to perform database operations such as querying and inserting documents.

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
		query := bson.D{{}}
		var employees []Employee = make([]Employee, 0)
		cursor, err := mg.Db.Collection("employees").Find(ctx.Context(), query)
		if err != nil {
			return ctx.Status(500).SendString(err.Error())
		}

		if err := cursor.All(ctx.Context(), &employees); err != nil {
			return ctx.Status(500).SendString(err.Error())
		}
		return ctx.JSON(employees)
	})

	app.Post("/employee", func(c *fiber.Ctx) error {
		collection := mg.Db.Collection("employees")
		employee := new(Employee)
		//	employee := ...: The result of new(Employee) is assigned to the variable employee.
		//	Since new returns a pointer to the allocated memory, employee is of type *Employee, which is a pointer to an Employee struct.

		employee.ID = ""
		insertionResult, err := collection.InsertOne(c.Context(), employee)
		if err != nil {
			return c.Status(500).SendString(err.Error())
		}

		// to ensure that the data is inserted, we will be using the inserted_ID to find and return that exact data
		filter := bson.D{{Key: "_id", Value: insertionResult.InsertedID}}
		createdRecord := collection.FindOne(c.Context(), filter)

		createdEmployee := &Employee{}
		err = createdRecord.Decode(createdEmployee)
		if err != nil {
			return err
		}

		return c.Status(201).JSON(createdEmployee)
	})
	app.Put("/employee/:id")
	app.Delete("/employee/:id")
}
