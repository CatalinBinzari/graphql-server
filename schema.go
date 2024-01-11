package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"

	"github.com/graphql-go/graphql"
)

// Helper function to import json from file to map
func importJSONDataFromFile(fileName string, result interface{}) (isOK bool) {
	isOK = true
	content, err := ioutil.ReadFile(fileName)
	if err != nil {
		fmt.Print("Error:", err)
		isOK = false
	}
	err = json.Unmarshal(content, result)
	if err != nil {
		isOK = false
		fmt.Print("Error:", err)
	}
	return
}

var BookList []Book
var _ = importJSONDataFromFile("./bookData", &BookList)

type Book struct {
	ID    int    `json:"bookId"`
	Name  string `json:"name"`
	Pages int    `json:"pages"`
}

// define custom GraphQL ObjectType `bookType` for our Golang struct `Book`
// Note that
// - the fields in our todoType maps with the json tags for the fields in our struct
// - the field type matches the field type in our struct
var bookType = graphql.NewObject(graphql.ObjectConfig{
	Name: "Book",
	Fields: graphql.Fields{
		"id": &graphql.Field{
			Type: graphql.Int,
		},
		"name": &graphql.Field{
			Type: graphql.String,
		},
		"pages": &graphql.Field{
			Type: graphql.Int,
		},
	},
})

var currentMaxId = 5

// root mutation
var rootMutation = graphql.NewObject(graphql.ObjectConfig{
	Name: "RootMutation",
	Fields: graphql.Fields{
		"addBook": &graphql.Field{
			Type:        bookType, // the return type for this field
			Description: "add a new book",
			Args: graphql.FieldConfigArgument{
				"name": &graphql.ArgumentConfig{
					Type: graphql.NewNonNull(graphql.String),
				},
				"pages": &graphql.ArgumentConfig{
					Type: graphql.NewNonNull(graphql.Int),
				},
			},
			Resolve: func(params graphql.ResolveParams) (interface{}, error) {

				// marshall and cast the argument value
				name, _ := params.Args["name"].(string)
				pages, _ := params.Args["pages"].(int)

				// figure out new id
				newID := currentMaxId + 1
				currentMaxId = currentMaxId + 1

				// perform mutation operation here
				// for e.g. create a Book and save to DB.
				newBook := Book{
					ID:    newID,
					Name:  name,
					Pages: pages,
				}

				BookList = append(BookList, newBook)

				// return the new Book object that we supposedly save to DB
				// Note here that
				// - we are returning a `Book` struct instance here
				// - we previously specified the return Type to be `bookType`
				// - `Book` struct maps to `bookType`, as defined in `bookType` ObjectConfig`
				return newBook, nil
			},
		},
		"updateBook": &graphql.Field{
			Type:        bookType, // the return type for this field
			Description: "Update existing book",
			Args: graphql.FieldConfigArgument{
				"name": &graphql.ArgumentConfig{
					Type: graphql.String,
				},
				"pages": &graphql.ArgumentConfig{
					Type: graphql.Int,
				},
				"id": &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.Int)},
			},
			Resolve: func(params graphql.ResolveParams) (interface{}, error) {
				id, _ := params.Args["id"].(int)
				affectedBook := Book{}

				// Search list for book with id
				for i := 0; i < len(BookList); i++ {
					if BookList[i].ID == id {
						if _, ok := params.Args["name"]; ok {
							BookList[i].Name = params.Args["name"].(string)
						}
						if _, ok := params.Args["pages"]; ok {
							BookList[i].Pages = params.Args["pages"].(int)
						}
						// Assign updated book so we can return it
						affectedBook = BookList[i]
						break
					}
				}
				// Return affected book
				return affectedBook, nil
			},
		},
	},
})

// root query
// test with Sandbox at localhost:8080/sandbox
var rootQuery = graphql.NewObject(graphql.ObjectConfig{
	Name: "RootQuery",
	Fields: graphql.Fields{
		"book": &graphql.Field{
			Type:        bookType,
			Description: "Get single book",
			Args: graphql.FieldConfigArgument{
				"name": &graphql.ArgumentConfig{
					Type: graphql.String,
				},
			},
			Resolve: func(params graphql.ResolveParams) (interface{}, error) {

				nameQuery, isOK := params.Args["name"].(string)
				if isOK {
					// Search for el with name
					for _, book := range BookList {
						if book.Name == nameQuery {
							return book, nil
						}
					}
				}

				return Book{}, nil
			},
		},

		"bookList": &graphql.Field{
			Type:        graphql.NewList(bookType),
			Description: "List of books",
			Resolve: func(p graphql.ResolveParams) (interface{}, error) {
				return BookList, nil
			},
		},
	},
})

// define schema, with our rootQuery and rootMutation
var BookSchema, _ = graphql.NewSchema(graphql.SchemaConfig{
	Query:    rootQuery,
	Mutation: rootMutation,
})
