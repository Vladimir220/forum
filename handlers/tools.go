package handlers

import (
	"encoding/json"
	"net/http"
)

func (h Handlers) readBodyReq(r *http.Request) (user user, err error) {
	contentType := r.Header.Get("Content-Type")
	switch contentType {
	case "application/json":
		err = json.NewDecoder(r.Body).Decode(&user)
		if err != nil {
			h.errorsLog.Printf("Ошибка кодирования в JSON: %v\n", err)
			return
		}
	case "application/x-www-form-urlencoded":
		err = r.ParseForm()
		if err != nil {
			h.errorsLog.Printf("Ошибка парсинга формы: %v\n", err)
			return
		}
		user.Username = r.FormValue("username")
		user.Password = r.FormValue("password")
	}
	return
}
