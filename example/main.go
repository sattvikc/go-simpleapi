package main

import (
	"fmt"

	"github.com/sattvikc/go-simpleapi"
	"github.com/sattvikc/go-simpleapi/middleware/cors"
	"github.com/sattvikc/go-simpleapi/middleware/logger"
)

type CreateBook struct {
	Body struct {
		Title  string `json:"title"`
		ISBN   string `json:"isbn"`
		Author string `json:"author"`
	} `body:"json"`
}

type CreateBookOK struct {
	Status string `json:"status"`
	Book   struct {
		Id     string `json:"id"`
		Title  string `json:"title"`
		ISBN   string `json:"isbn"`
		Author string `json:"author"`
	} `json:"book"`
}

type CreateBookExists struct {
	Status string `json:"status"`
	Book   struct {
		Id     string `json:"id"`
		Title  string `json:"title"`
		ISBN   string `json:"isbn"`
		Author string `json:"author"`
	} `json:"book"`
}

func createBook(e *simpleapi.Endpoint) interface{} {
	e.WithTag("Books").
		WithResponse(200, CreateBookOK{}, "Book created").
		WithResponse(200, CreateBookExists{}, "Book already exists").
		POST()

	return func(ctx *simpleapi.Context, req CreateBook) error {
		fmt.Printf("Request: %+v", req)

		ctx.JSON(200, map[string]interface{}{
			"status": "OK",
		})

		return nil
	}
}

func withAuth(e *simpleapi.Endpoint) interface{} {
	type Unauthorised struct {
		Reason string `json:"reason"`
	}

	e.WithResponse(401, Unauthorised{}, "Unauthorised")

	return func(ctx *simpleapi.Context, headers struct {
		Authorization string `header:"Authorization"`
	}) error {
		fmt.Println("Authorization:", headers.Authorization)
		return ctx.JSON(401, Unauthorised{
			Reason: "Token expired",
		})
	}
}

func main() {
	app := simpleapi.New()

	app.Use(cors.New())
	app.Use(logger.New())

	app.Endpoint("/books", withAuth, createBook)

	app.ListenAndServe(":8000")
}
