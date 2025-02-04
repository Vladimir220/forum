package DAO

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"forum/crypto"
	sm "forum/system_models"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"
)

type DaoPostgres struct {
	db *sql.DB
}

// Слушает обновления комментариев к посту.
func (dp DaoPostgres) ListenPostComments(ctx context.Context, postID string) chan *sm.Comment {
	commentsChan := make(chan *sm.Comment)

	go func() {
		defer close(commentsChan)

		lastCommentID := "0"

		for {
			select {
			case <-ctx.Done():
				return
			default:
				rows, err := dp.db.QueryContext(ctx, `SELECT id, content, author_id, post_id, parent_id 
														FROM comments 
														WHERE post_id = $1 AND id > $2 
														ORDER BY id ASC`, postID, lastCommentID)
				if err != nil {
					return
				}
				defer rows.Close()

				for rows.Next() {
					var comment sm.Comment
					var authorId int
					if err := rows.Scan(&comment.ID, &comment.Content, &authorId, &comment.PostID, &comment.ParentId); err != nil {
						return
					}

					author, err := dp.ReadUserByID(strconv.Itoa(authorId))
					if err != nil {
						return
					}
					comment.Author = author.Username

					select {
					case <-ctx.Done():
						return
					default:
						commentsChan <- &comment
						lastCommentID = comment.ID
					}
				}

				time.Sleep(2 * time.Second)
			}
		}
	}()

	return commentsChan
}

// Обновляет доступность поста для комментирования.
func (dp DaoPostgres) UpdateCommentingAccess(username string, postId string, access bool) (err error) {
	if postId == "" || username == "" {
		err = errors.New("postId или username не указано")
		return
	}

	post, err := dp.ReadPostByID(postId)
	if err != nil {
		return
	}
	if post.Author != username {
		err = fmt.Errorf("нет прав на изменения поста ID-%s пользователем %s", postId, username)
		return
	}

	queryStr := "UPDATE posts SET allow_comments = $1 WHERE id = $2;"

	_, err = dp.db.Exec(queryStr, access, postId)

	return
}

// Возвращает пост по его ID.
func (dp DaoPostgres) ReadPostByID(postId string) (post sm.Post, err error) {
	if postId == "" {
		err = errors.New("postId не указано")
		return
	}

	queryStr := "SELECT p.id, title, content, allow_comments, user_name FROM posts p INNER JOIN users u ON p.author_id = u.id  WHERE p.id=$1;"

	err = dp.db.QueryRow(queryStr, postId).Scan(&post.ID, &post.Title, &post.Content, &post.AllowComments, &post.Author)
	if err != nil {
		err = fmt.Errorf("ошибка при поиске поста по id: %v", err)
		return
	}

	return
}

// Читает комментарии ближайшего уровня вложенности, соответствующие указанным ID постов и ID родительских комментариев (контекстам комментариев).
//
// Создано для загрузчика.
func (dp DaoPostgres) ReadNearCommentsByCtx(commentCtx []sm.CommentCtx, limit, offset int) (commentsForEveryCtx map[sm.CommentCtx][]*sm.Comment, err error) {
	commentsForEveryCtx = map[sm.CommentCtx][]*sm.Comment{}
	postIdVals := []string{}
	i := 1
	for ; i <= len(commentCtx); i++ {
		postIdVals = append(postIdVals, fmt.Sprintf("$%d", i))
	}
	parentIdVals := []string{}
	for ; i <= len(commentCtx)*2; i++ {
		parentIdVals = append(parentIdVals, fmt.Sprintf("$%d", i))
	}

	args := []any{}
	for _, v := range commentCtx {
		args = append(args, v.PostId)
	}
	for _, v := range commentCtx {
		if v.ParentId == "" {
			args = append(args, nil)
		} else {
			args = append(args, v.ParentId)
		}
	}

	var queryStr string
	var rows *sql.Rows
	if limit == -1 {
		queryStr = fmt.Sprintf(`WITH BufComments AS (
								SELECT comments.id, content, user_name, post_id, parent_id,
								ROW_NUMBER() OVER (PARTITION BY comments.post_id ORDER BY comments.id ASC) as rn
								FROM comments INNER JOIN users ON comments.author_id = users.id
								WHERE post_id IN (%s) AND (parent_id IN (%s) OR (parent_id IS NULL AND (%s) IS NULL))
								)
								SELECT id, content, user_name, post_id, parent_id
								FROM BufComments
								WHERE rn > %d;`, strings.Join(postIdVals, ", "), strings.Join(parentIdVals, ", "), strings.Join(parentIdVals, ", "), offset)
		rows, err = dp.db.Query(queryStr, args...)
	} else {
		queryStr = fmt.Sprintf(`WITH BufComments AS (
								SELECT comments.id, content, user_name, post_id, parent_id,
								ROW_NUMBER() OVER (PARTITION BY comments.post_id ORDER BY comments.id ASC) as rn
								FROM comments INNER JOIN users ON comments.author_id = users.id
								WHERE post_id IN (%s) AND (parent_id IN (%s) OR (parent_id IS NULL AND (%s) IS NULL))
								)
								SELECT id, content, user_name, post_id, parent_id
								FROM BufComments
								WHERE rn > %d AND rn <= %d;`, strings.Join(postIdVals, ", "), strings.Join(parentIdVals, ", "), strings.Join(parentIdVals, ", "), offset, limit)
		rows, err = dp.db.Query(queryStr, args...)
	}
	if err != nil {
		err = fmt.Errorf("ошибка при загрузки ближайших комментариев по контектсу: %v", err)
		return
	}

	defer rows.Close()

	for rows.Next() {
		var comm sm.Comment
		err = rows.Scan(&comm.ID, &comm.Content, &comm.Author, &comm.PostID, &comm.ParentId)
		var commCtx sm.CommentCtx
		if comm.ParentId == nil {
			commCtx.ParentId = ""
		} else {
			commCtx.ParentId = *comm.ParentId
		}
		commCtx.PostId = comm.PostID

		_, ok := commentsForEveryCtx[commCtx]
		if !ok {
			commentsForEveryCtx[commCtx] = []*sm.Comment{}
		}
		commentsForEveryCtx[commCtx] = append(commentsForEveryCtx[commCtx], &comm)

		if err != nil {
			err = fmt.Errorf("ошибка получения результатов: %v", err)
			return
		}
	}

	return
}

// Возвращает список всёх постов.
func (dp DaoPostgres) ReadAllPosts(limit, offset int) (posts []*sm.Post, err error) {
	var queryStr string
	var rows *sql.Rows
	if limit == -1 {
		queryStr = `SELECT posts.id, title, content, allow_comments, user_name
					FROM posts INNER JOIN users ON posts.author_id = users.id 
					OFFSET $1;`
		rows, err = dp.db.Query(queryStr, offset)
	} else {
		queryStr = `SELECT posts.id, title, content, allow_comments, user_name
					FROM posts INNER JOIN users ON posts.author_id = users.id 
					OFFSET $1 LIMIT $2;`
		rows, err = dp.db.Query(queryStr, offset, limit)
	}

	if err != nil {
		err = fmt.Errorf("ошибка при чтении списка постов: %v", err)
		return
	}
	defer rows.Close()

	for rows.Next() {
		post := &sm.Post{}
		err = rows.Scan(&post.ID, &post.Title, &post.Content, &post.AllowComments, &post.Author)
		if err != nil {
			err = fmt.Errorf("ошибка получения результатов: %v", err)
			return
		}
		posts = append(posts, post)
	}

	return
}

// Создаёт новый комментарий.
func (dp DaoPostgres) CreateComment(author sm.User, postID string, parentID *string, content string) (comment sm.Comment, err error) {
	if author.ID == "" || author.Username == "" || postID == "" || content == "" {
		err = errors.New("или author.ID, или author.Username, или postID, или content не указано")
		return
	}

	var queryStr string
	if parentID == nil {
		queryStr = `INSERT INTO comments (content, author_id, post_id) 
					SELECT $1, $2, $3
					WHERE EXISTS (
						SELECT 1 FROM posts 
						WHERE id = $3 AND allow_comments = true
					)
					RETURNING id, content, parent_id, post_id;`
		err = dp.db.QueryRow(queryStr, content, author.ID, postID).Scan(&comment.ID, &comment.Content, &comment.ParentId, &comment.PostID)
	} else {
		queryStr = `INSERT INTO comments (content, author_id, post_id, parent_id) 
					SELECT ($1, $2, $3, $4)
					WHERE EXISTS (
						SELECT 1 FROM posts 
						WHERE id = $3 AND allow_comments = true
					)
					RETURNING id, content, parent_id, post_id;`
		comment.ParentId = new(string)
		err = dp.db.QueryRow(queryStr, content, author.ID, postID, parentID).Scan(&comment.ID, &comment.Content, comment.ParentId, &comment.PostID)
	}

	if err != nil {
		err = fmt.Errorf("ошибка при создании нового комментария: %v", err)
		return
	}

	comment.Author = author.Username
	return
}

// Создаёт новый пост.
func (dp DaoPostgres) CreatePost(author sm.User, title, content string, allowComments bool) (post sm.Post, err error) {
	if author.ID == "" || author.Username == "" || title == "" || content == "" {
		err = errors.New("или author.ID, или author.Username, или title, или content не указано")
		return
	}

	queryStr := `INSERT INTO posts (title, content, allow_comments, author_id) 
				VALUES ($1, $2, $3, $4) 
				RETURNING id, title, content, allow_comments;`

	err = dp.db.QueryRow(queryStr, title, content, allowComments, author.ID).Scan(&post.ID, &post.Title, &post.Content, &post.AllowComments)
	if err != nil {
		err = fmt.Errorf("ошибка при создании нового поста: %v", err)
		return
	}
	post.Author = author.Username
	return
}

// Возвращает данные пользователя, включая захэшированный пароль, по его username.
func (dp DaoPostgres) ReadUserDataByName(username string) (user sm.UserDb, err error) {
	if username == "" {
		err = errors.New("username не указано")
		return
	}

	queryStr := "SELECT * FROM users WHERE user_name=$1;"

	err = dp.db.QueryRow(queryStr, username).Scan(&user.ID, &user.Username, &user.HashedPassword)
	if err != nil {
		err = fmt.Errorf("ошибка при поиске данных пользователя по username: %v", err)
		return
	}
	return
}

// Возвращает пользователя по его ID.
func (dp DaoPostgres) ReadUserByID(userId string) (user sm.User, err error) {
	if userId == "" {
		err = errors.New("userId не указано")
		return
	}

	queryStr := "SELECT * FROM users WHERE id=$1;"

	var pass string
	err = dp.db.QueryRow(queryStr, userId).Scan(&user.ID, &user.Username, &pass)
	if err != nil {
		err = fmt.Errorf("ошибка при поиске пользователя по id: %v", err)
		return
	}

	return
}

// Создаёт нового пользователя.
func (dp DaoPostgres) CreateUser(name, password string) (id int, err error) {
	if name == "" || password == "" {
		err = errors.New("name или password не указано")
		return
	}

	queryStr := "INSERT INTO users (user_name, hashed_password) VALUES ($1, $2) RETURNING id;"

	hashedPassword, err := crypto.GetHashedPassword(password)
	if err != nil {
		err = fmt.Errorf("ошибка при хэшировании пароля: %v", err)
		return
	}

	err = dp.db.QueryRow(queryStr, name, hashedPassword).Scan(&id)
	if err != nil {
		err = fmt.Errorf("ошибка при создании нового пользователя: %v", err)
		return
	}
	return
}

func (dp *DaoPostgres) init() (err error) {
	var (
		user     = os.Getenv("DB_USER")
		password = os.Getenv("DB_PASSWORD")
		dbName   = os.Getenv("DB_NAME")
		host     = os.Getenv("DB_HOST")
	)
	if user == "" || password == "" || dbName == "" || host == "" {
		err = errors.New("в env не указана одна из следующих переменных: DB_USER, DB_PASSWORD, DB_NAME, DB_HOST")
		return
	}

	loginInfo := fmt.Sprintf("postgres://%s:%s@%s/%s?sslmode=disable", user, password, host, "postgres")

	db, err := sql.Open("postgres", loginInfo)
	if err != nil {
		err = fmt.Errorf("ошибка подключения к БД: %v", err)
		return
	}

	query := "SELECT EXISTS (SELECT datname FROM pg_catalog.pg_database WHERE datname = $1)"

	var isDbExist bool
	err = db.QueryRow(query, dbName).Scan(&isDbExist)
	if err != nil {
		err = fmt.Errorf("ошибка при проверке существования базы данных: %v", err)
		return
	}

	if !isDbExist {
		_, err = db.Exec(fmt.Sprintf("CREATE DATABASE %s;", dbName))
		if err != nil {
			err = fmt.Errorf("ошибка при создании базы данных: %v", err)
			return
		}
	}

	loginInfo = fmt.Sprintf("postgres://%s:%s@%s/%s?sslmode=disable", user, password, host, dbName)

	dp.db, err = sql.Open("postgres", loginInfo)
	if err != nil {
		err = fmt.Errorf("ошибка подключения к БД: %v", err)
		return
	}

	return
}

func (dao DaoPostgres) checkMigrations() (err error) {
	driver, err := postgres.WithInstance(dao.db, &postgres.Config{})
	if err != nil {
		err = fmt.Errorf("ошибка создания драйвера: %v", err)
		return
	}

	tmod := os.Getenv("TEST_MOD")
	dbName := os.Getenv("DB_NAME")
	var path string
	if tmod != "1" {
		path = ""
	} else {
		currentDir, _ := os.Getwd()
		path = filepath.ToSlash(filepath.Dir(currentDir))
		if path != "" {
			path = path + "/"
		}
	}

	m, err := migrate.NewWithDatabaseInstance("file://"+path+"db/migration", dbName, driver)
	if err != nil {
		err = fmt.Errorf("ошибка создания мигратора: %v", err)
		return
	}

	if err = m.Up(); err != nil && err != migrate.ErrNoChange {
		err = fmt.Errorf("ошибка применения миграций: %v", err)
		return
	} else {
		err = nil
	}

	return
}

// Закрывает соединение с базой данных.
func (dao *DaoPostgres) Close() (err error) {
	err = dao.db.Close()
	if err != nil {
		err = fmt.Errorf("ошибка закрытия БД: %v", err)
	}
	return
}

// Создаёт и инициализирует DAO Postgres, проверяет миграции.
//
// Обязательно нужны следующие переменные env: DB_HOST, DB_USER, DB_PASSWORD, DB_NAME.
//
// Миграции должны лежать тут: "./db/migration".
func CreateDaoPostgres() (dao *DaoPostgres, err error) {
	psql := &DaoPostgres{}
	err = psql.init()
	if err != nil {
		return
	}

	err = psql.checkMigrations()
	if err != nil {
		return
	}

	return psql, err
}
