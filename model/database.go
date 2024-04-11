package model

import (
	"fmt"
	"time"
)

type Database interface {
	ConnStr()
}

type DatabaseConnection struct {
	User     string
	Password string
	Timeout  time.Duration
	Host     string
	Port     string
}

func (d DatabaseConnection) ConnStr() string {
	if d.User == "" || d.Password == "" {
		return fmt.Sprintf("mongodb://%s:%s", d.Host, d.Port)
	} else {
		return fmt.Sprintf("mongodb://%s:%s@%s:%s", d.User, d.Password, d.Host, d.Port)
	}
}
