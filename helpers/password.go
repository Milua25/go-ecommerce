package helpers

import (
	"golang.org/x/crypto/bcrypt"
)

var passwordCost = 14

func HashPassword(password string) (string, error) {

	// Hash the password using bcrypt with a cost of 14
	strBytes, err := bcrypt.GenerateFromPassword([]byte(password), passwordCost)
	if err != nil {
		return "", err
	}
	return string(strBytes), nil
}

func VerifyPassword(hashedPassword, password string) (bool, string) {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	if err != nil {
		return false, "Incorrect password"
	}
	return true, ""
}
