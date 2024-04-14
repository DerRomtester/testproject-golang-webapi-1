package model

import (
	"time"

	"github.com/google/uuid"
)

type Session interface {
	IsExpired() bool
	RenewSession(t time.Duration) (UserSession, string)
}

type UserSession struct {
	Username string    `json:"username"`
	Expiry   time.Time `json:"expiry"`
}

func (s UserSession) IsExpired() bool {
	return s.Expiry.Before(time.Now())
}

func (s UserSession) RenewSession(t time.Duration) (UserSession, string) {
	newSessionToken := uuid.NewString()
	expiresAt := time.Now().Add(t * time.Second)

	return UserSession{
		Username: s.Username,
		Expiry:   expiresAt,
	}, newSessionToken
}
