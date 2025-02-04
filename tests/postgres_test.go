package tests

import (
	"forum/db/DAO"
	"testing"
	"time"
)

func TestPostgresPost(t *testing.T) {
	envInit()

	db, err := DAO.CreateDaoPostgres()
	if err != nil {
		t.Fatal(err)
	}

	time.Sleep(time.Second * 2)
	UserTest(db, t)
	PostTest(db, t)
}
