package auth

import "golang.org/x/crypto/bcrypt"

func generatePassword(p string) string {
	hash, _ := bcrypt.GenerateFromPassword([]byte(p), bcrypt.DefaultCost)
	return string(hash)
}

func comparePassword(hashedPassword, password string) bool {
	// Make sure we don't allow empty passwords since accounts using providers may not have a password.
	if password == "" {
		return false
	}
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	return err == nil
}
