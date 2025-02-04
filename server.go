package main

import (
	"fmt"
	"forum/db/DAO"
	"forum/inits"
	"net/http"
	"os"
)

const defaultHost = "localhost:1234"

func main() {

	errorsLog, dbErrorsLog := inits.InitSystem()

	host := os.Getenv("API_HOST")
	if host == "" {
		host = defaultHost
	}

	activeDB := os.Getenv("ACTIVE_DB")
	var dao DAO.Dao
	var err error
	if activeDB == "redis" {
		dao, err = DAO.CreateDaoRedis()
	} else if activeDB == "postgres" {
		dao, err = DAO.CreateDaoPostgres()
	} else {
		dao, err = DAO.CreateDaoPostgres()
	}
	defer dao.Close()

	if err != nil {
		dbErrorsLog.Fatal(err)
	}

	router := inits.InitRouter(dao, errorsLog, dbErrorsLog)

	fmt.Printf("\nХост: %s/\n", host)

	err = http.ListenAndServe(host, router)
	if err != nil {
		errorsLog.Fatal(err)
	}
}
