package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/DerRomtester/testproject-golang-webapi-1/database"
	"github.com/DerRomtester/testproject-golang-webapi-1/handler"
	"github.com/DerRomtester/testproject-golang-webapi-1/model"

	"go.mongodb.org/mongo-driver/mongo"
)

var (
	err    error
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

func main() {
	client = InitDB()
	mux := http.NewServeMux()
	port = ":8080"

	mux.HandleFunc("PUT /auth", func(w http.ResponseWriter, r *http.Request) {
		handler.HandlePutLogout(w, r, client)
	})

	mux.HandleFunc("POST /auth", func(w http.ResponseWriter, r *http.Request) {
		err = handler.HandlePostLogin(w, r)
		if err != nil {
			fmt.Println(err)
		}
	})

	mux.HandleFunc("GET /devices", func(w http.ResponseWriter, r *http.Request) {
		handler.HandleGetDevices(w, r, client)
	})

	mux.HandleFunc("POST /devices", func(w http.ResponseWriter, r *http.Request) {
		handler.HandlePostDevices(w, r, client)
	})

	mux.HandleFunc("DELETE /devices", func(w http.ResponseWriter, r *http.Request) {
		handler.HandleDeleteDevices(w, r, client)
	})

	mux.HandleFunc("GET /device/{id}", func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		handler.HandleGetDeviceByID(w, r, client, id)
	})

	mux.HandleFunc("DELETE /device/{id}", func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		handler.HandleDeleteDevice(w, r, client, id)
	})

	mux.HandleFunc("GET /session", func(w http.ResponseWriter, r *http.Request) {
		handler.HandleGetSession(w, r)
	})

	mux.HandleFunc("PUT /refresh", func(w http.ResponseWriter, r *http.Request) {
		handler.HandlePutRefreshToken(w, r)
	})

	log.Printf("Starting server on port %s\n", port)

	err = http.ListenAndServe(port, mux)
	if err != nil {
		log.Fatal(err)
	}
}
