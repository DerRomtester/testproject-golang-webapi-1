package main

import (
	"log"
	"net/http"
	"time"

	"github.com/DerRomtester/testproject-golang-webapi-1/database"
	"github.com/DerRomtester/testproject-golang-webapi-1/handler"
	"github.com/DerRomtester/testproject-golang-webapi-1/model"

	"go.mongodb.org/mongo-driver/mongo"
)

type Server struct {
	Host string
	Port string
}

var (
	db = model.DatabaseConnection{
		Host:    "mongo",
		Port:    "27017",
		Timeout: 30 * time.Second,
	}
	srv = Server{
		Host: "localhost",
		Port: ":8080",
	}
)

func InitDB() *mongo.Client {
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
		handler.HandlePostLogin(w, r)
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
		handler.HandleGetSession(w, r)
	})

	mux.HandleFunc("PUT /v1/refresh", func(w http.ResponseWriter, r *http.Request) {
		handler.HandlePutRefreshToken(w, r)
	})

	return mux
}

func main() {
	client := InitDB()
	mux := buildHandlers(client)

	log.Printf("Starting server on port %s\n", srv.Port)
	err := http.ListenAndServe(srv.Port, mux)
	if err != nil {
		log.Fatal(err)
	}
}
