package graph

import (
	"forum/db/DAO"
	"log"
)

//go:generate go run github.com/99designs/gqlgen generate

// This file will not be regenerated automatically.
//
// It serves as dependency injection for your app, add any dependencies you require here.

type Resolver struct {
	Dao         DAO.Dao
	ErrorsLog   *log.Logger
	DbErrorsLog *log.Logger
}
