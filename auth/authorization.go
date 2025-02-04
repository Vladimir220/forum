package auth

import (
	"context"
	"errors"
	"forum/crypto"
	"forum/db/DAO"
	sm "forum/system_models"
	"net/http"
	"time"
)

var userCtxKey = сontextKey{"user"}

type сontextKey struct {
	name string
}

// Авторизация пользователей.
func AuthMiddleware(db DAO.Dao) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			c, err := r.Cookie("auth-cookie")

			if err != nil || c == nil {
				next.ServeHTTP(w, r)
				return
			}

			userId, err := crypto.ValidateToken(c.Value)
			if err != nil {
				http.SetCookie(w, &http.Cookie{
					Name:    "auth-cookie",
					Expires: time.Now().Add(-1 * time.Hour),
				})
				http.Error(w, "Требуется повторная аутентификация!", http.StatusUnauthorized)
				return
			}

			user, err := db.ReadUserByID(userId)
			if err != nil {
				http.SetCookie(w, &http.Cookie{
					Name:    "auth-cookie",
					Expires: time.Now().Add(-1 * time.Hour),
				})

				http.Error(w, "Подозрительный токен!", http.StatusForbidden)
				return
			}

			ctx := context.WithValue(r.Context(), userCtxKey, user)

			r = r.WithContext(ctx)
			next.ServeHTTP(w, r)
		})
	}
}

// Получение авторизованного пользователя из контекста
func GetUserFromContext(ctx context.Context) (user sm.User, err error) {
	ctxValue := ctx.Value(userCtxKey)
	if ctxValue == nil {
		err = errors.New("не авторизован")
		return
	}
	user = ctxValue.(sm.User)
	return
}
