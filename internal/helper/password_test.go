package helper_test

import (
	"math/rand"
	"testing"

	"github.com/DerRomtester/testproject-golang-webapi-1/internal/helper"
	"github.com/stretchr/testify/assert"
)

func GenerateRandomPassword() string {
	const randString = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ123456789"
	b := make([]byte, 10)
	for i := range b {
		b[i] = randString[rand.Intn((len(randString)))]
	}
	return string(b)
}

func TestHashPassword(t *testing.T) {
	password := GenerateRandomPassword()
	hashedPassword, err := helper.HashPassword(password)
	assert.Nil(t, err, "HashPassword should not return an error")
	assert.NotEmpty(t, hashedPassword, "Hashed password should not be empty")
}

func TestComparePasswordandHash(t *testing.T) {
	password := GenerateRandomPassword()
	hashedPassword, err := helper.HashPassword(password)

	assert.Nil(t, err, "hashed password should not return an error")
	err = helper.CheckPassword(password, hashedPassword)
	assert.Nil(t, err, "CheckPassword should not return an error with correct password")

	wrongPassword := GenerateRandomPassword()
	err = helper.CheckPassword(wrongPassword, hashedPassword)
	assert.Errorf(t, err, "CheckPassword should return error with incorrect password")
}
