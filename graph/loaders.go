package graph

/*
// import vikstrous/dataloadgen with your other imports
import (
	"context"
	"database/sql"
	"forum/graph/model"
	"net/http"
	"strings"
	"time"

	"github.com/vikstrous/dataloadgen"
)

type ctxKey string

const (
	loadersKey = ctxKey("dataloaders")
)

// userReader reads Users from a database
type userReader struct {
	db *sql.DB
}

// getUsers implements a batch function that can retrieve many users by ID,
// for use in a dataloader
func (u *userReader) getUsers(ctx context.Context, userIDs []string) ([]*model.User, []error) {
	stmt, err := u.db.PrepareContext(ctx, `SELECT id, name FROM users WHERE id IN (?`+strings.Repeat(",?", len(userIDs)-1)+`)`)
	if err != nil {
		return nil, []error{err}
	}
	defer stmt.Close()

	rows, err := stmt.QueryContext(ctx, userIDs)
	if err != nil {
		return nil, []error{err}
	}
	defer rows.Close()

	users := make([]*model.User, 0, len(userIDs))
	errs := make([]error, 0, len(userIDs))
	for rows.Next() {
		var user model.User
		err := rows.Scan(&user.ID, &user.Name)
		users = append(users, &user)
		errs = append(errs, err)
	}
	return users, errs
}

// Loaders wrap your data loaders to inject via middleware
type Loaders struct {
	UserLoader *dataloadgen.Loader[string, *model.User]
}

// NewLoaders instantiates data loaders for the middleware
func NewLoaders(conn *sql.DB) *Loaders {
	// define the data loader
	ur := &userReader{db: conn}
	return &Loaders{
		UserLoader: dataloadgen.NewLoader(ur.getUsers, dataloadgen.WithWait(time.Millisecond)),
	}
}

// Middleware injects data loaders into the context
func Middleware(conn *sql.DB, next http.Handler) http.Handler {
	// return a middleware that injects the loader to the request context
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		loader := NewLoaders(conn)
		r = r.WithContext(context.WithValue(r.Context(), loadersKey, loader))
		next.ServeHTTP(w, r)
	})
}

// For returns the dataloader for a given context
func For(ctx context.Context) *Loaders {
	return ctx.Value(loadersKey).(*Loaders)
}

// GetUser returns single user by id efficiently
func GetUser(ctx context.Context, userID string) (*model.User, error) {
	loaders := For(ctx)
	return loaders.UserLoader.Load(ctx, userID)
}

// GetUsers returns many users by ids efficiently
func GetUsers(ctx context.Context, userIDs []string) ([]*model.User, error) {
	loaders := For(ctx)
	return loaders.UserLoader.LoadAll(ctx, userIDs)
}
*/
