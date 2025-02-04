package handlers

import (
	"encoding/json"
	"net/http"
)

func readBodyReq(w http.ResponseWriter, r *http.Request) (user user, err error) {
	contentType := r.Header.Get("Content-Type")
	switch contentType {
	case "application/json":
		err = json.NewDecoder(r.Body).Decode(&user)
		if err != nil {
			http.Error(w, "Failed to decode JSON", http.StatusBadRequest)
			return
		}
	case "application/x-www-form-urlencoded":
		err = r.ParseForm()
		if err != nil {
			http.Error(w, "Failed to parse form", http.StatusBadRequest)
			return
		}
		user.Username = r.FormValue("username")
		user.Password = r.FormValue("password")
	}
	return
}
