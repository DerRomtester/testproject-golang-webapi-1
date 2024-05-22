package main

import (
	"fmt"
	"log"
	"time"

	"github.com/DerRomtester/testproject-golang-webapi-1/handler"
	"github.com/DerRomtester/testproject-golang-webapi-1/internal/database"

	"github.com/spf13/viper"
)

func GetConfig() (handler.ServerConfig, database.DatabaseConnection) {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AutomaticEnv()

	err := viper.ReadInConfig()
	if err != nil {
		fmt.Println("error config file: default \n", err)
	}
	db := database.DatabaseConnection{
		User:     viper.GetString("DatabaseConnection.User"),
		Password: viper.GetString("DatabaseConnection.Password"),
		Host:     viper.GetString("DatabaseConnection.Host"),
		Port:     viper.GetString("DatabaseConnection.Port"),
		Timeout:  viper.GetDuration("DatabaseConnection.Timeout") * time.Second,
	}

	srv := handler.ServerConfig{
		Domain: viper.GetString("Server.Domain"),
		Port:   viper.GetString("Server.Port"),
	}
	return srv, db
}

func main() {
	srv, db := GetConfig()
	client, err := db.ConnectDB()

	if err != nil {
		log.Fatal(err)
	}

	srv.Run(client)

	if err != nil {
		log.Fatal(err)
	}
}
