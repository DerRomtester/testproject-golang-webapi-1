package database_test

import (
	"testing"

	"github.com/DerRomtester/testproject-golang-webapi-1/internal/database"
	"github.com/magiconair/properties/assert"
)

var (
	dbUserPw = database.DatabaseConnection{
		User:     "TestUser",
		Password: "TestPassword",
		Timeout:  600,
		Host:     "localhost",
		Port:     "27015",
	}

	dbNoUserPw = database.DatabaseConnection{
		Timeout: 600,
		Host:    "localhost",
		Port:    "27015",
	}
)

func TestGetConnectionStringUserPassword(t *testing.T) {
	connString := dbUserPw.GetConnStr()
	expected := "mongodb://TestUser:TestPassword@localhost:27015"
	assert.Equal(t, connString, expected, "connectionString does not equal mongodb://Testuser:TestPasswords@localhost:27015")
}

func TestGetConnectionStringNoUserPassword(t *testing.T) {
	connString := dbNoUserPw.GetConnStr()
	expected := "mongodb://localhost:27015"
	assert.Equal(t, connString, expected, "connectionString does not equal mongodb://localhost:27015")
}
