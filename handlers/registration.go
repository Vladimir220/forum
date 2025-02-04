package handlers

import (
	"forum/crypto"
	"net/http"
	"strconv"
	"time"
)

func (h Handlers) Registration(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Маршрут поддерживает только POST", http.StatusMethodNotAllowed)
		return
	}

	user, err := readBodyReq(w, r)

	if err != nil {
		http.Error(w, "Ожидаются поля username и password", http.StatusBadRequest)
		return
	}

	userId, err := h.daoDB.CreateUser(user.Username, user.Password)

	// Баги не будут останавливать работу сервера, но мы их логируем
	if err != nil {
		h.dbErrorsLog.Println(err)
		http.Error(w, "Ошибка сервера", http.StatusInternalServerError)
		return
	}

	token, err := crypto.GenerateToken(strconv.Itoa(userId))
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
