package tests

import "os"

func envInit() {
	// не ожидается запуск тестов внутри контейнера
	os.Setenv("DB_HOST", "localhost:5432")
	os.Setenv("DB_USER", "postgres")
	os.Setenv("DB_PASSWORD", "qwerty")
	os.Setenv("DB_NAME", "forum")

	os.Setenv("REDIS_HOST", "localhost:6379")
	os.Setenv("REDIS_PASSWORD", "")
	os.Setenv("REDIS_DB", "0")

	os.Setenv("TEST_MOD", "1")
}
