package DAO

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"forum/crypto"
	sm "forum/system_models"
	"os"
	"strconv"
	"strings"

	"github.com/redis/go-redis/v9"
)

const (
	userIDCounter    = "user:id:counter"
	postIDCounter    = "post:id:counter"
	commentIDCounter = "comment:id:counter"
)

type DaoRedis struct {
	rdb *redis.Client
	ctx context.Context
}

// Слушает обновления комментариев к посту.
func (dp DaoRedis) ListenPostComments(ctx context.Context, postID string) chan *sm.Comment {
	ch := make(chan *sm.Comment)

	go func() {
		defer close(ch)
		pubsub := dp.rdb.Subscribe(ctx, "sub:comment:postid:"+postID)

		defer pubsub.Close()

		for {
			msg, err := pubsub.ReceiveMessage(ctx)
			if err != nil {
				return
			}

			var comment sm.Comment
			err = json.Unmarshal([]byte(msg.Payload), &comment)
			if err != nil {
				return
			}

			select {
			case <-ctx.Done():
				return
			case ch <- &comment:
			}
		}
	}()

	return ch
}

// Возвращает пост по его ID.
func (dr DaoRedis) ReadPostByID(postId string) (post sm.Post, err error) {
	if postId == "" {
		err = errors.New("postId не указано")
		return
	}

	postKey := fmt.Sprintf("post:id:%s", postId)

	data, err := dr.rdb.HGetAll(dr.ctx, postKey).Result()
	if err != nil {
		err = fmt.Errorf("ошибка при поиске поста по id: %v", err)
		return
	}

	post.ID = postId
	post.Title = data["title"]
	post.Content = data["content"]
	post.AllowComments, _ = strconv.ParseBool(data["allowComments"])
	post.Author = data["author"]

	return
}

// Обновляет доступность поста для комментирования.
func (dr DaoRedis) UpdateCommentingAccess(username string, postId string, access bool) (err error) {
	if postId == "" || username == "" {
		err = errors.New("postId или username не указано")
		return
	}

	post, err := dr.ReadPostByID(postId)
	if err != nil {
		return
	}
	if post.Author != username {
		err = fmt.Errorf("нет прав на изменения поста ID-%s пользователем %s", postId, username)
		return
	}

	postKey := fmt.Sprintf("post:id:%s", postId)
	err = dr.rdb.HSet(dr.ctx, postKey, "allowComments", access).Err()
	return
}

// Создаёт нового пользователя.
func (dr DaoRedis) CreateUser(name, password string) (id int, err error) {
	if name == "" || password == "" {
		err = errors.New("name или password не указано")
		return
	}

	idIncr, err := dr.rdb.Incr(dr.ctx, userIDCounter).Result()
	if err != nil {
		err = fmt.Errorf("ошибка при создании нового пользователя: %v", err)
		return
	}

	hashedPassword, err := crypto.GetHashedPassword(password)
	if err != nil {
		err = fmt.Errorf("ошибка при хэшировании пароля: %v", err)
		return
	}

	userKey := fmt.Sprintf("user:id:%d", idIncr)
	err = dr.rdb.SetNX(dr.ctx, userKey, name, 0).Err()
	if err != nil {
		err = fmt.Errorf("ошибка при создании нового пользователя: %v", err)
		return
	}

	userKey = fmt.Sprintf("user:name:%s", name)

	err = dr.rdb.HSet(dr.ctx, userKey, "id", idIncr, "password", hashedPassword).Err()
	if err != nil {
		err = fmt.Errorf("ошибка при создании нового пользователя: %v", err)
		return
	}
	id = int(idIncr)
	return
}

// Возвращает пользователя по его ID.
func (dr DaoRedis) ReadUserByID(userId string) (user sm.User, err error) {
	if userId == "" {
		err = errors.New("userId не указано")
		return
	}

	userKey := fmt.Sprintf("user:id:%s", userId)
	username, err := dr.rdb.Get(dr.ctx, userKey).Result()
	if err != nil {
		err = fmt.Errorf("ошибка при поиске пользователя по id: %v", err)
		return
	}

	user.ID = userId
	user.Username = username

	return
}

// Возвращает данные пользователя, включая захэшированный пароль, по его username.
func (dr DaoRedis) ReadUserDataByName(username string) (user sm.UserDb, err error) {
	if username == "" {
		err = errors.New("username не указано")
		return
	}

	userKey := fmt.Sprintf("user:name:%s", username)

	data, err := dr.rdb.HGetAll(dr.ctx, userKey).Result()
	if err != nil || len(data) == 0 {
		err = fmt.Errorf("ошибка при поиске данных пользователя по username: %v", err)
		return
	}

	user.ID = data["id"]
	user.Username = username
	user.HashedPassword = data["password"]

	return
}

// Создаёт новый пост.
func (dr DaoRedis) CreatePost(author sm.User, title, content string, allowComments bool) (post sm.Post, err error) {
	if author.ID == "" || author.Username == "" || title == "" || content == "" {
		err = errors.New("или author.ID, или author.Username, или title, или content не указано")
		return
	}

	idIncr, err := dr.rdb.Incr(dr.ctx, postIDCounter).Result()
	if err != nil {
		err = fmt.Errorf("ошибка при создании нового поста: %v", err)
		return
	}

	postKey := fmt.Sprintf("post:id:%d", idIncr)
	err = dr.rdb.HSet(dr.ctx, postKey, "title", title, "content", content, "allowComments", allowComments, "author", author.Username).Err()
	if err != nil {
		err = fmt.Errorf("ошибка при создании нового поста: %v", err)
		return
	}

	post.ID = fmt.Sprintf("%d", idIncr)
	post.Title = title
	post.Content = content
	post.AllowComments = allowComments
	post.Author = author.Username

	return
}

// Возвращает список всёх постов.
func (dr DaoRedis) ReadAllPosts(limit, offset int) (posts []*sm.Post, err error) {
	keys, err := dr.rdb.Keys(dr.ctx, "post:id:*").Result()
	if err != nil {
		return nil, err
	}

	var max int
	if limit == -1 || offset+limit+1 > len(keys) {
		max = len(keys)
	} else {
		max = offset + limit
	}

	var min int
	if offset > len(keys) {
		min = len(keys)
	}

	for _, key := range keys[min:max] {
		if key == "post:id:counter" {
			continue
		}
		post, err := dr.ReadPostByID(strings.TrimPrefix(key, "post:id:"))

		if err != nil {
			err = fmt.Errorf("ошибка при чтении списка постов: %v", err)
			return nil, err
		}
		posts = append(posts, &post)
	}

	return
}

// Создаёт новый комментарий.
func (dr DaoRedis) CreateComment(author sm.User, postID string, parentID *string, content string) (comment sm.Comment, err error) {
	if author.ID == "" || author.Username == "" || postID == "" || content == "" {
		err = errors.New("или author.ID, или author.Username, или postID, или content не указано")
		return
	}

	post, err := dr.ReadPostByID(postID)
	if err != nil {
		err = fmt.Errorf("ошибка при создании нового комментария: %v", err)
		return
	}

	if !post.AllowComments {
		err = fmt.Errorf("комментарии отключены")
		return
	}

	idIncr, err := dr.rdb.Incr(dr.ctx, commentIDCounter).Result()
	if err != nil {
		err = fmt.Errorf("ошибка при создании нового комментария: %v", err)
		return
	}

	commentKey := fmt.Sprintf("comment:postid:%s:id:%d", postID, idIncr)
	err = dr.rdb.HSet(dr.ctx, commentKey, "content", content, "author", author.Username).Err()
	if err != nil {
		err = fmt.Errorf("ошибка при создании нового комментария: %v", err)
		return
	}

	parentIDStr := ""
	if parentID != nil {
		parentIDStr = *parentID
	}

	commentKey = fmt.Sprintf("comment:postid:%s:parentid:%s", postID, parentIDStr)
	err = dr.rdb.LPush(dr.ctx, commentKey, idIncr).Err()
	if err != nil {
		err = fmt.Errorf("ошибка при создании нового комментария: %v", err)
		return
	}

	comment.ID = fmt.Sprintf("%d", idIncr)
	comment.PostID = postID
	comment.Content = content
	comment.Author = author.Username
	comment.ParentId = parentID

	commentJSON, err := json.Marshal(comment)
	if err != nil {
		err = fmt.Errorf("ошибка при упаковке комментария в JSON: %v", err)
		return
	}

	err = dr.rdb.Publish(dr.ctx, "sub:comment:postid:"+postID, commentJSON).Err()
	if err != nil {
		err = fmt.Errorf("ошибка при публикации комментария: %v", err)
		return
	}

	return
}

// Читает комментарии ближайшего уровня вложенности, соответствующие указанным ID постов и ID родительских комментариев (контекстам комментариев).
//
// Создано для загрузчика.
func (dr DaoRedis) ReadNearCommentsByCtx(commentCtx []sm.CommentCtx, limit, offset int) (commentsForEveryCtx map[sm.CommentCtx][]*sm.Comment, err error) {
	commentsForEveryCtx = map[sm.CommentCtx][]*sm.Comment{}

	for _, v := range commentCtx {
		commentKey := fmt.Sprintf("comment:postid:%s:parentid:%s", v.PostId, v.ParentId)

		ids, err := dr.rdb.LRange(dr.ctx, commentKey, int64(offset), int64(limit)).Result()
		if err != nil {
			err = fmt.Errorf("ошибка при загрузки ближайших комментариев по контектсу: %v", err)
			return commentsForEveryCtx, err
		}

		for _, id := range ids {
			commentKey = fmt.Sprintf("comment:postid:%s:id:%s", v.PostId, id)

			data, err := dr.rdb.HGetAll(dr.ctx, commentKey).Result()
			if err != nil {
				err = fmt.Errorf("ошибка получения результатов: %v", err)
				return commentsForEveryCtx, err
			}

			var comm sm.Comment
			comm.ID = id
			comm.PostID = v.PostId
			comm.ParentId = &v.ParentId
			comm.Content = data["content"]
			comm.Author = data["author"]

			_, ok := commentsForEveryCtx[v]
			if !ok {
				commentsForEveryCtx[v] = []*sm.Comment{}
			}
			commentsForEveryCtx[v] = append(commentsForEveryCtx[v], &comm)
		}
	}

	return
}

func (r *DaoRedis) Init() (err error) {
	var (
		host     = os.Getenv("REDIS_HOST")
		password = os.Getenv("REDIS_PASSWORD")
		dbStr    = os.Getenv("REDIS_DB")
	)

	if host == "" || dbStr == "" {
		err = errors.New("в env не указана одна из следующих переменных: REDIS_HOST, REDIS_PASSWORD, REDIS_DB")
		return
	}
	db, err := strconv.Atoi(dbStr)
	if err != nil {
		return
	}

	r.rdb = redis.NewClient(&redis.Options{
		Addr:     host,
		Password: password,
		DB:       db,
	})

	_, err = r.rdb.Ping(context.Background()).Result()
	if err != nil {
		err = fmt.Errorf("не удалось подключиться к Redis: %v", err)
		return
	}
	r.ctx = context.Background()

	r.rdb.FlushDB(r.ctx)

	return
}

// Закрывает соединение с базой данных.
func (dao *DaoRedis) Close() (err error) {
	err = dao.rdb.Close()
	if err != nil {
		err = fmt.Errorf("ошибка закрытия БД: %v", err)
	}
	return
}

func CreateDaoRedis() (dao *DaoRedis, err error) {
	dao = &DaoRedis{}
	err = dao.Init()
	return
}
