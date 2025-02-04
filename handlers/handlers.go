package handlers

import (
	"forum/db/DAO"
	"log"
)

// Содержит обработчики для маршрутов.
type Handlers struct {
	daoDB       DAO.Dao
	errorsLog   *log.Logger
	dbErrorsLog *log.Logger
}

// Создаёт и инициализирует обработчики для маршрутов.
func CreateHandlers(daoDB DAO.Dao, errorsLog *log.Logger, dbErrorsLog *log.Logger) (h Handlers) {
	h = Handlers{daoDB: daoDB, errorsLog: errorsLog, dbErrorsLog: dbErrorsLog}
	return
}
