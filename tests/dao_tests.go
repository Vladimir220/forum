package tests

import (
	"forum/crypto"
	"forum/db/DAO"
	"forum/system_models"
	"strconv"
	"testing"
)

func PostTest(dao DAO.Dao, t *testing.T) {
	username, _ := generateRandomString(5)
	userId, err := dao.CreateUser(username, "qwerty")
	if err != nil {
		t.Fatalf("Ошибка работы dao.CreateUser: %v", err)
	}
	user := system_models.User{ID: strconv.Itoa(userId), Username: username}

	post1, err := dao.CreatePost(user, "название", "содержимое", true)
	if err != nil {
		t.Fatalf("Ошибка работы dao.CreatePost: %v", err)
	}

	post2, err := dao.ReadPostByID(post1.ID)
	if err != nil {
		t.Fatalf("Ошибка работы dao.ReadPostByID: %v", err)
	}

	if post1 != post2 {
		t.Fatal("Ошибка работы dao.ReadPostByID")
	}

	posts, err := dao.ReadAllPosts(-1, 0)
	if err != nil {
		t.Fatalf("Ошибка работы dao.ReadAllPosts: %v", err)
	}
	len1 := len(posts)

	_, err = dao.CreatePost(user, "название", "содержимое", true)
	if err != nil {
		t.Fatalf("Ошибка работы dao.CreatePost: %v", err)
	}

	posts, err = dao.ReadAllPosts(-1, 0)
	if err != nil {
		t.Fatalf("Ошибка работы dao.ReadAllPosts: %v", err)
	}
	len2 := len(posts)

	if (len1 + 1) != len2 {
		t.Fatal("Ошибка работы dao.ReadAllPosts")
	}
}

func UserTest(dao DAO.Dao, t *testing.T) {
	username, _ := generateRandomString(5)
	userId, err := dao.CreateUser(username, "qwerty")
	if err != nil {
		t.Fatalf("Ошибка работы dao.CreateUser: %v", err)
	}

	user1 := system_models.User{ID: strconv.Itoa(userId), Username: username}
	user2, err := dao.ReadUserByID(strconv.Itoa(userId))
	if err != nil {
		t.Fatalf("Ошибка работы dao.ReadUserByID: %v", err)
	}
	if user1 != user2 {
		t.Fatal("Ошибка работы dao.ReadUserByID")
	}

	user3, err := dao.ReadUserDataByName(username)
	if err != nil {
		t.Fatalf("Ошибка работы dao.ReadUserDataByName: %v", err)
	}
	if user1.ID != user3.ID {
		t.Fatal("Ошибка работы dao.ReadUserDataByName")
	}

	isPassEqls := crypto.ComparePassword("qwerty", user3.HashedPassword)
	if !isPassEqls {
		t.Fatal("Ошибка работы патека crypto")
	}
}
