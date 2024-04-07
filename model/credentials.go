package model

type Credentials struct {
	Password string `json:"password"`
	Username string `json:"username"`
}

type MongDBConnection struct {
	User     string
	Password string
	Timeout  string
	Host     string
	Port     string
}
