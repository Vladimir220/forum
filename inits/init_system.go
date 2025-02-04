package inits

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

// Читает файлы env, открывает файлы log.
func InitSystem() (errorsLog *log.Logger, dbErrorsLog *log.Logger) {
	err := godotenv.Load("./env/conf.env", "./env/.env")
	if err != nil {
		err = godotenv.Load("./env/.env")
		if err != nil {
			panic(err.Error())
		}
	}

	errorsLogFile, err := os.OpenFile(os.Getenv("ERRORS_LOG_FILE_PATH"), os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		panic(err.Error())
	}
	dbErrorsLogFile, err := os.OpenFile(os.Getenv("DB_ERRORS_LOG_FILE_PATH"), os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		panic(err.Error())
	}

	errorsLog = log.New(errorsLogFile, "ERROR:", log.Ldate|log.Ltime|log.Lshortfile|log.Lmsgprefix)
	dbErrorsLog = log.New(dbErrorsLogFile, "DB_ERROR:", log.Ldate|log.Ltime|log.Lshortfile|log.Lmsgprefix)
	return
}
