package tests

import (
	"forum/db/DAO"
	"testing"
)

func TestPostgresPost(t *testing.T) {
	envInit()

	db, err := DAO.CreateDaoPostgres()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	UserTest(db, t)
	PostTest(db, t)
	CommentTest(db, t)
}
