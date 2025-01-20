package utils

import "golang.org/x/crypto/bcrypt"

func GeneratePassword(p string) string {
	hash, _ := bcrypt.GenerateFromPassword([]byte(p), bcrypt.DefaultCost)
	return string(hash)
}

func ComparePassword(hashedPassword string, password string) bool {
	// Make sure we don't allow empty passwords since accounts using OAuth providers may not have a password
	if password == "" {
		return false
	}
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	return err == nil
}
