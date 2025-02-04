package auth

import (
	"forum/crypto"
	"forum/db/DAO"
)

// Аутентификация пользователей.
func AuthenticateUser(username, password string, dao DAO.Dao) (userId string, exist bool) {
	exist = true

	user, err := dao.ReadUserDataByName(username)
	if err != nil {
		exist = false
		return
	}
	userId = user.ID

	exist = crypto.ComparePassword(password, user.HashedPassword)
	return
}
