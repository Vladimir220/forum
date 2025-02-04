package loaders

import (
	"context"
	"forum/db/DAO"
	"log"
	"net/http"
)

func CommentsLoaderMiddleware(dao DAO.Dao, dbErrorsLog *log.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			loader := &commentsLoaders{dao: dao, dbErrorsLog: dbErrorsLog}
			loader.init()

			r = r.WithContext(context.WithValue(r.Context(), loaderKey, loader))
			next.ServeHTTP(w, r)

		})
	}
}
