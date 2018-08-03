package main

import (
	"fmt"
	"time"

	"github.com/kataras/iris/context"
	"github.com/kataras/iris/middleware/logger"
	"github.com/kataras/iris/middleware/recover"

	"github.com/kataras/iris"
	mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

// User Struct
type User struct {
	ID         bson.ObjectId `bson:"_id,omitempty"`
	Firsname   string        `json:"firstname"`
	Lastname   string        `json:"lastname"`
	Age        int           `json:"age"`
	Msisdn     string        `json:"msisdn"`
	InsertedAt time.Time     `json:"inserted_at" bson:"isnerted_at"`
	LastUpdate time.Time     `json:"last_update" bson:"last_update"`
}

func main() {
	app := iris.New()
	app.Logger().SetLevel("debug")
	app.Use(recover.New())
	app.Use(logger.New())

	// Connect Mongo DB
	session, err := mgo.Dial("localhost")
	if err != nil {
		panic(err)
	}

	defer session.Close()
	session.SetMode(mgo.Monotonic, true)

	c := session.DB("usergoapi").C("profiles")

	// Index
	index := mgo.Index{
		Key:        []string{"msisdn"},
		Unique:     true,
		DropDups:   true,
		Background: true,
		Sparse:     true,
	}

	err = c.EnsureIndex(index)

	if err != nil {
		panic(err)
	}

	// Endpoint default
	app.Handle("GET", "/", func(ctx context.Context) {
		ctx.JSON(context.Map{"message": "Welcome to Simple API"})
	})

	// Method: POST
	// Create new user

	app.Handle("POST", "/users", func(ctx context.Context) {
		params := &User{}
		err := ctx.ReadJSON(params)
		if err != nil {
			ctx.JSON(context.Map{"response": err.Error()})
		} else {
			params.LastUpdate = time.Now()
			params.InsertedAt = time.Now()
			err := c.Insert(params)
			if err != nil {
				ctx.JSON(context.Map{"response": err.Error()})
			} else {
				fmt.Println("Successfully inserted data")
				result := User{}
				err := c.Find(bson.M{"msisdn": params.Msisdn}).One(&result)
				if err != nil {
					ctx.JSON(context.Map{"response": err.Error()})
				}
				ctx.JSON(context.Map{"response": "User successfully created", "data": result})
			}
		}
	})

	// Method: Get
	// Get All Users
	app.Handle("GET", "/users", func(ctx context.Context) {
		results := []User{}

		err := c.Find(nil).All(&results)
		if err != nil {
			ctx.JSON(context.Map{"response": "No data found"})
		} else {
			ctx.JSON(context.Map{"response": "Success process the request", "data": results})
		}
	})

	// Method: Get
	// Get Single User
	app.Handle("GET", "/users/{msisdn:string}", func(ctx context.Context) {
		msisdn := ctx.Params().Get("msisdn")
		if msisdn == "" {
			ctx.JSON(context.Map{"response": "Please pass a valid msisdn"})
		}

		result := User{}
		err := c.Find(bson.M{"msisdn": msisdn}).One(&result)

		if err != nil {
			ctx.JSON(context.Map{"response": err.Error()})
		} else {
			ctx.JSON(context.Map{"response": "Success process the request", "data": result})
		}

	})

	// Method: Pacth
	// This is update user
	app.Handle("PATCH", "/users/{msisdn: string}", func(ctx context.Context) {
		msisdn := ctx.Params().Get("msisdn")
		fmt.Println(msisdn)
		if msisdn == "" {
			ctx.JSON(context.Map{"response": "Please pass a valid msisdn"})
		}
		params := &User{}
		err := ctx.ReadJSON(params)
		if err != nil {
			ctx.JSON(context.Map{"response": err.Error()})
		} else {
			params.LastUpdate = time.Now()
			params.InsertedAt = time.Now()
			query := bson.M{"msisdn": msisdn}
			err = c.Update(query, params)
			if err != nil {
				ctx.JSON(context.Map{"response": err.Error()})
			} else {
				result := User{}
				err = c.Find(bson.M{"msisdn": params.Msisdn}).One(&result)
				if err != nil {
					ctx.JSON(context.Map{"response": err.Error()})
				}
				ctx.JSON(context.Map{"response": "user record successfully updated", "data": result})
			}
		}

	})

	// Method: Delete
	// Delete user
	app.Handle("DELETE", "/users/{msisdn: string}", func(ctx context.Context) {
		msisdn := ctx.Params().Get("msisdn")
		if msisdn == "" {
			ctx.JSON(context.Map{"response": "Please pass a valid msisdn"})
		}
		params := &User{}
		err := ctx.ReadJSON(params)
		if err != nil {
			ctx.JSON(context.Map{"response": err.Error()})
		} else {
			query := bson.M{"msisdn": msisdn}
			err = c.Remove(query)
			if err != nil {
				ctx.JSON(context.Map{"response": err.Error()})
			} else {
				ctx.JSON(context.Map{"response": "User record successfully deleted"})
			}
		}
	})

	app.Run(iris.Addr(":8080"), iris.WithoutServerError(iris.ErrServerClosed))

}
