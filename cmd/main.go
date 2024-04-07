package main

import (
	"log"
	"net/http"

	"github.com/DerRomtester/testproject-golang-webapi-1/handler"

	"go.mongodb.org/mongo-driver/mongo"
)

var (
	client *mongo.Client
	err    error
)

func main() {
	mux := http.NewServeMux()

	mux.HandleFunc("PUT /auth", func(w http.ResponseWriter, r *http.Request) {
		handler.HandlePutLogout(w, r, client)
	})

	mux.HandleFunc("POST /auth", func(w http.ResponseWriter, r *http.Request) {
		client, err = handler.HandlePostLogin(w, r)
		if err != nil {
			log.Fatal(err)
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

	err = http.ListenAndServe(":8080", mux)
	if err != nil {
		log.Fatal(err)
	}
}
