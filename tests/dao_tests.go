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

	//dao.UpdateCommentingAccess()
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

func CommentTest(dao DAO.Dao, t *testing.T) {
	username, _ := generateRandomString(5)
	userId, err := dao.CreateUser(username, "qwerty")
	if err != nil {
		t.Fatalf("Ошибка работы dao.CreateUser: %v", err)
	}
	user := system_models.User{ID: strconv.Itoa(userId), Username: username}

	post, err := dao.CreatePost(user, "название", "содержимое", true)
	if err != nil {
		t.Fatalf("Ошибка работы dao.CreatePost: %v", err)
	}
	potsId := post.ID

	comment1 := system_models.Comment{
		PostID:   potsId,
		ParentId: nil,
		Content:  "qwerty",
		Author:   user.Username,
	}
	comment2, err := dao.CreateComment(user, potsId, nil, "qwerty")
	if err != nil {
		t.Fatalf("Ошибка работы dao.CreateComment: %v", err)
	}
	if comment1.ParentId != comment2.ParentId || comment1.PostID != comment2.PostID || comment1.Author != comment2.Author || comment1.Content != comment2.Content {
		t.Fatal("Ошибка работы dao.CreateComment")
	}
	parentId := comment2.ID
	comment3 := system_models.Comment{
		PostID:   potsId,
		ParentId: &parentId,
		Content:  "qwerty",
		Author:   user.Username,
	}
	comment4, err := dao.CreateComment(user, potsId, &parentId, "qwerty")
	if err != nil {
		t.Fatalf("Ошибка работы dao.CreateComment: %v", err)
	}
	if *(comment3.ParentId) != *(comment4.ParentId) || comment3.PostID != comment4.PostID || comment3.Author != comment4.Author || comment3.Content != comment4.Content {
		t.Fatal("Ошибка работы dao.CreateComment")
	}

	comCtx := system_models.CommentCtx{
		PostId:   potsId,
		ParentId: comment2.ID,
	}
	comCtxs := []system_models.CommentCtx{comCtx}

	commentsForEveryCtx, err := dao.ReadNearCommentsByCtx(comCtxs, 1, 0)
	if err != nil {
		t.Fatalf("Ошибка работы dao.ReadNearCommentsByCtx: %v", err)
	}
	arrRes, ok := commentsForEveryCtx[comCtx]
	if !ok || len(arrRes) == 0 {
		t.Fatal("Ошибка работы dao.ReadNearCommentsByCtx")
	}
	comment5 := *(arrRes[0])
	if *(comment4.ParentId) != *(comment5.ParentId) || comment4.PostID != comment5.PostID || comment4.Author != comment5.Author || comment4.Content != comment5.Content || comment4.ID != comment5.ID {
		t.Fatal("Ошибка работы dao.CreateComment")
	}

}
