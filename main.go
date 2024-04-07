package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
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
const mongoURI = "mongodb://localhost:27017/" + dbName

type Employee struct {
	ID     string  `json:"id,omitempty" bson:"_id,omitempty"`
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
		query := bson.D{{}} //creates an empty BSON document (bson.D{{}}) to use as a query filter.

		var employees []Employee = make([]Employee, 0)
		cursor, err := mg.Db.Collection("employees").Find(ctx.Context(), query)
		//Then, it initializes an empty slice of Employee structs ([]Employee) using the make function.
		//After that, it executes a query to retrieve all documents from the "employees" collection in the MongoDB database.
		//The result is a cursor (cursor) that can be iterated over to access each document.

		if err != nil {
			return ctx.Status(500).SendString(err.Error())
		}

		if err := cursor.All(ctx.Context(), &employees); err != nil {
			return ctx.Status(500).SendString(err.Error())
		}
		//If there's no error, it reads all the documents from the cursor into the employees slice using the All() method.
		//If an error occurs during this process, it again sets the HTTP status code to 500 and sends the error message as the response body.
		return ctx.JSON(employees)
	})

	app.Post("/employee", func(c *fiber.Ctx) error {
		collection := mg.Db.Collection("employees")
		employee := new(Employee)
		//	employee := ...: The result of new(Employee) is assigned to the variable employee.
		//	Since new returns a pointer to the allocated memory, employee is of type *Employee, which is a pointer to an Employee struct.

		if err := c.BodyParser(employee); err != nil {
			//c.BodyParser(employee); This line attempts to parse the request body of the current HTTP request (c). The BodyParser function within Fiber tries to decode
			//the request body based on the content type headers and populate the fields of the provided employee struct with the parsed data (assuming the request body is JSON formatted).
			return c.Status(400).SendString(err.Error())
		}

		employee.ID = ""
		insertionResult, err := collection.InsertOne(c.Context(), employee)
		if err != nil {
			return c.Status(500).SendString(err.Error())
		}

		// to ensure that the data is inserted, we will be using the inserted_ID to find and return that exact data
		//bson.D: This is a type from the MongoDB driver used to create a Document (BSON document).
		filter := bson.D{{Key: "_id", Value: insertionResult.InsertedID}}
		createdRecord := collection.FindOne(c.Context(), filter)
		//This line attempts to find a single document from the "employees" collection that matches the filter criteria (filter).

		createdEmployee := &Employee{}
		//A new Employee struct pointer (createdEmployee) is created to store the retrieved document.

		err = createdRecord.Decode(createdEmployee)
		//createdRecord.Decode decodes the retrieved document (if found) into the createdEmployee struct.
		//This method decodes the retrieved BSON document into the Go struct format (Employee).
		if err != nil {
			return err
		}

		return c.Status(201).JSON(createdEmployee)
	})

	app.Put("/employee/:id", func(c *fiber.Ctx) error {
		idParam := c.Params("id")

		employeeId, err := primitive.ObjectIDFromHex(idParam)
		log.Println("employee Id", employeeId)
		//employeeId, err := primitive.ObjectIDFromHex(idParam): This attempts to convert the idParam string (which should be a hexadecimal ObjectID string) into a MongoDB ObjectID type (primitive.ObjectID).
		if err != nil {
			return c.SendStatus(400)
		}
		employee := new(Employee)

		if err := c.BodyParser(employee); err != nil {
			return c.Status(400).SendString(err.Error())
		}
		log.Println(employee)

		query := bson.D{{Key: "_id", Value: employeeId}}
		update := bson.D{
			{Key: "$set",
				Value: bson.D{
					{Key: "name", Value: employee.Name},
					{Key: "age", Value: employee.Age},
					{Key: "salary", Value: employee.Salary},
				},
			},
		}
		err = mg.Db.Collection("employees").FindOneAndUpdate(c.Context(), query, update).Err()
		if err != nil {
			if errors.Is(err, mongo.ErrNoDocuments) {
				//Here, errors.Is(err, mongo.ErrNoDocuments) checks if the err is an instance of mongo.ErrNoDocuments or wraps an instance of it.
				//This approach ensures compatibility with various error types and avoids issues related to comparing errors directly with equality operators.
				return c.SendStatus(400)
			}
			return c.SendStatus(500)
		}
		employee.ID = idParam
		//The code assigns idParam to employee.ID again for a few key reasons:
		//1. Consistency in Response:
		//
		//The employee struct is sent back to the client as a JSON response. Without this assignment, the ID field in the response might be empty or have a default value, even though the record was updated successfully in the database.
		//	Reassigning the idParam to employee.ID ensures that the response accurately reflects the current state of the employee in the database, including its correct ID.
		//
		//	2. Avoiding ObjectID Type Issues:
		//
		//	The primitive.ObjectIDFromHex function converts the string-based idParam into a MongoDB-specific primitive.ObjectID type.
		//	Directly using employeeId (the ObjectID type) in JSON encoding might lead to compatibility issues with certain JSON libraries or clients.
		//	Reassigning the original string-based idParam back to employee.ID guarantees a consistent string representation of the ID in the response, ensuring compatibility with various clients.

		return c.Status(200).JSON(employee)
	})

	app.Delete("/employee/:id", func(c *fiber.Ctx) error {
		employeeId, err := primitive.ObjectIDFromHex(c.Params("id"))
		if err != nil {
			return c.SendStatus(400)
		}
		query := bson.D{{Key: "_id", Value: employeeId}}
		result, err := mg.Db.Collection("employees").DeleteOne(c.Context(), query)
		if err != nil {
			return c.SendStatus(500)
		}
		if result.DeletedCount < 1 {
			return c.SendStatus(404)
		}

		return c.Status(201).JSON("Record Deleted Successfully.")
	})

	log.Fatal(app.Listen(":3000"))
	//	if there is error while running server it will exit
}
