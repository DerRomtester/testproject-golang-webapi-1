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

var (
	port   string
	client *mongo.Client

	db = model.DatabaseConnection{
		Host:    "localhost",
		Port:    "27017",
		Timeout: 5 * time.Second,
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
	port := ":8080"
	mux := buildHandlers(client)

	log.Printf("Starting server on port %s\n", port)
	err := http.ListenAndServe(port, mux)
	if err != nil {
		log.Fatal(err)
	}
}
