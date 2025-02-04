package crypto

import "golang.org/x/crypto/bcrypt"

// Хэширует пароль.
func GetHashedPassword(password string) (hashedPassword string, err error) {
	res, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return
	}
	hashedPassword = string(res)
	return
}

// Проверяет на соответствие пароль и его возможно захэшированную версию.
func ComparePassword(password, hashedPassword string) (equal bool) {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	if err == nil {
		equal = true
	}
	return
}
