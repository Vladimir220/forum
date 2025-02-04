package DAO

import (
	"context"
	sm "forum/system_models"
)

type Dao interface {
	// Слушает обновления комментариев к посту.
	ListenPostComments(ctx context.Context, postID string) chan *sm.Comment

	// Возвращает пост по его ID.
	ReadPostByID(postId string) (post sm.Post, err error)

	// Обновляет доступность поста для комментирования.
	UpdateCommentingAccess(username string, postId string, access bool) (err error)

	// Создаёт нового пользователя.
	CreateUser(name, password string) (id int, err error)

	// Возвращает пользователя по его ID.
	ReadUserByID(userId string) (user sm.User, err error)

	// Возвращает данные пользователя, включая захэшированный пароль, по его username.
	ReadUserDataByName(username string) (user sm.UserDb, err error)

	// Создаёт новый пост.
	CreatePost(author sm.User, title, content string, allowComments bool) (post sm.Post, err error)

	// Возвращает список всёх постов.
	ReadAllPosts(limit, offset int) (posts []*sm.Post, err error)

	// Создаёт новый комментарий.
	CreateComment(author sm.User, postID string, parentID *string, content string) (comment sm.Comment, err error)

	// Читает комментарии ближайшего уровня вложенности, соответствующие указанным ID постов и ID родительских комментариев (контекстам комментариев).
	//
	// Создано для загрузчика.
	ReadNearCommentsByCtx(commentCtx []sm.CommentCtx, limit, offset int) (commentsForEveryCtx map[sm.CommentCtx][]*sm.Comment, err error)

	// Закрывает соединение с базой данных.
	Close() (err error)
}
