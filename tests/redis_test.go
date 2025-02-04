package tests

import (
	"forum/db/DAO"
	"testing"
)

func TestRedisPost(t *testing.T) {
	envInit()

	db, err := DAO.CreateDaoRedis()
	if err != nil {
		t.Fatal(err)
	}

	UserTest(db, t)
	PostTest(db, t)
}
