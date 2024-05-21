package session_test

import (
	"testing"
	"time"

	"github.com/DerRomtester/testproject-golang-webapi-1/internal/session"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

var (
	testExpiry = time.Now().Add(0 * time.Second)
	testToken  = uuid.NewString()
	testUser   = "TestUser"

	s = session.UserSession{
		Username: testUser,
		Expiry:   testExpiry,
		Token:    testToken,
	}
)

func TestRenewSession(t *testing.T) {
	newSession := s.RenewSession(3600)
	assert.NotEqual(t, testExpiry, newSession.Expiry, "the expiry did not change")
}

func TestGetToken(t *testing.T) {
	token := s.GetToken()
	assert.Equal(t, token, &testToken, "the tokens do no match")
}
