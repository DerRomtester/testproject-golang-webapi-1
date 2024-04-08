package model

type Database interface {
}

type DatabaseConnection struct {
	User     string
	Password string
	Timeout  string
	Host     string
	Port     string
}

func (d DatabaseConnection) ConnStr() string {
	if d.User == "" || d.Password == "" {
		return "mongodb://" + d.Host + ":" + d.Port
	} else {
		return "mongodb://" + d.User + ":" + d.Password + "@" + d.Host + ":" + d.Port
	}
}
