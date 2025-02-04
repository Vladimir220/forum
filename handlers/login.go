package handlers

import (
	"forum/auth"
	"forum/crypto"
	"net/http"
	"time"
)

func (h Handlers) Login(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Маршрут поддерживает только POST", http.StatusMethodNotAllowed)
		return
	}

	user, err := readBodyReq(w, r)

	if err != nil {
		http.Error(w, "Ожидаются поля username и password", http.StatusBadRequest)
		return
	}

	userId, exist := auth.AuthenticateUser(user.Username, user.Password, h.daoDB)
	if !exist {
		http.Error(w, "Неправильные username или password", http.StatusUnauthorized)
		return
	}

	token, err := crypto.GenerateToken(userId)
	if err != nil {
		h.errorsLog.Println(err)
		http.Error(w, "Ошибка сервера", http.StatusInternalServerError)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "auth-cookie",
		Value:    token,
		HttpOnly: true,
		Expires:  time.Now().Add(crypto.GetExpirationHours()),
	})
}
