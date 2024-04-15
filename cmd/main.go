package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/DerRomtester/testproject-golang-webapi-1/database"
	"github.com/DerRomtester/testproject-golang-webapi-1/handler"
	"github.com/DerRomtester/testproject-golang-webapi-1/model"

	"github.com/spf13/viper"
	"go.mongodb.org/mongo-driver/mongo"
)

type Server struct {
	Host string
	Port string
}

func GetConfig() (Server, model.DatabaseConnection) {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AutomaticEnv()

	err := viper.ReadInConfig()
	if err != nil {
		fmt.Println("error config file: default \n", err)
	}
	db := model.DatabaseConnection{
		User:     viper.GetString("DatabaseConnection.User"),
		Password: viper.GetString("DatabaseConnection.Password"),
		Host:     viper.GetString("DatabaseConnection.Host"),
		Port:     viper.GetString("DatabaseConnection.Port"),
		Timeout:  viper.GetDuration("DatabaseConnection.Timeout") * time.Second,
	}

	srv := Server{
		Host: viper.GetString("Server.Host"),
		Port: viper.GetString("Server.Port"),
	}
	return srv, db
}

func InitDB(db model.DatabaseConnection) *mongo.Client {
	c, err := database.ConnectDB(db)

	if err != nil {
		log.Fatal(err)
	}

	return c
}

func buildHandlers(client *mongo.Client) *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("PUT /v1/auth", func(w http.ResponseWriter, r *http.Request) {
		handler.HandlePutLogout(w, r, client)
	})

	mux.HandleFunc("POST /v1/auth", func(w http.ResponseWriter, r *http.Request) {
		handler.HandlePostLogin(w, r, client)
	})

	mux.HandleFunc("GET /v1/devices", func(w http.ResponseWriter, r *http.Request) {
		handler.HandleGetDevices(w, r, client)
	})

	mux.HandleFunc("POST /v1/devices", func(w http.ResponseWriter, r *http.Request) {
		handler.HandlePostDevices(w, r, client)
	})

	mux.HandleFunc("DELETE /v1/devices", func(w http.ResponseWriter, r *http.Request) {
		handler.HandleDeleteDevices(w, r, client)
	})

	mux.HandleFunc("GET /v1/device/{id}", func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		handler.HandleGetDeviceByID(w, r, client, id)
	})

	mux.HandleFunc("DELETE /v1/device/{id}", func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		handler.HandleDeleteDevice(w, r, client, id)
	})

	mux.HandleFunc("GET /v1/session", func(w http.ResponseWriter, r *http.Request) {
		handler.HandleGetSession(w, r, client)
	})

	mux.HandleFunc("PUT /v1/refresh", func(w http.ResponseWriter, r *http.Request) {
		handler.HandlePutRefreshToken(w, r, client)
	})

	mux.HandleFunc("POST /v1/user", func(w http.ResponseWriter, r *http.Request) {
		handler.HandleCreateUser(w, r, client)
	})

	return mux
}

func main() {
	srv, db := GetConfig()
	client := InitDB(db)
	mux := buildHandlers(client)

	log.Printf("Starting server on port %s\n", srv.Port)
	err := http.ListenAndServe(srv.Port, mux)
	if err != nil {
		log.Fatal(err)
	}
}
