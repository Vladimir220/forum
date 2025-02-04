package handlers

import (
	"net/http"
	"os"
)

func (h Handlers) LoginForm(w http.ResponseWriter, r *http.Request) {
	htmlFile, err := os.ReadFile("./html/login.html")
	if err != nil {
		http.Error(w, "Не удалось загрузить страницу", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	w.Write(htmlFile)
}
