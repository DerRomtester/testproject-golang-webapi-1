package session

import (
	"time"

	"github.com/google/uuid"
)

type UserSession struct {
	Username string    `json:"username"`
	Expiry   time.Time `json:"expiry"`
	Token    string
}

func (s UserSession) GetSession() *UserSession {
	return &UserSession{}
}

func (s UserSession) GetToken() *string {
	return &s.Token
}

func (s UserSession) IsExpired() bool {
	return s.Expiry.Before(time.Now())
}

func (s UserSession) RenewSession(t time.Duration) UserSession {
	newSessionToken := uuid.NewString()
	expiresAt := time.Now().Add(t * time.Second)

	return UserSession{
		Username: s.Username,
		Expiry:   expiresAt,
		Token:    newSessionToken,
	}
}
