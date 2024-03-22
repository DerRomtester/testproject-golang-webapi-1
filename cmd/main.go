package main

import (
	"github.com/DerRomtester/testproject-golang-webapi-1/handler"
	"github.com/DerRomtester/testproject-golang-webapi-1/model"
	"log"
	"net/http"

	"go.mongodb.org/mongo-driver/mongo"
)

const (
	MethodGet    = "GET"
	MethodPost   = "POST"
	MethodPut    = "PUT"
	MethodDelete = "DELETE"
)

var (
	client *mongo.Client
)

func main() {
	var APIMethodNotAllowed model.APIError
	APIMethodNotAllowed.Err = "method not allowed"

	http.HandleFunc("/auth", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case MethodPut:
			handler.HandlePutLogout(w, r, client)
		case MethodPost:
			client = handler.HandlePostLogin(w, r)
		default:
			handler.HTTPJsonMsg(w, APIMethodNotAllowed, http.StatusMethodNotAllowed)
		}
	})

	http.HandleFunc("/devices", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case MethodGet:
			handler.HandleGetDevices(w, r, client)
		case MethodPost:
			handler.HandlePostDevices(w, r, client)
		case MethodDelete:
			handler.HandleDeleteDevices(w, r, client)
		default:
			handler.HTTPJsonMsg(w, APIMethodNotAllowed, http.StatusMethodNotAllowed)
		}
	})

	http.HandleFunc("/device", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case MethodGet:
			handler.HandleGetDeviceByID(w, r, client)
		case MethodDelete:
			handler.HandleDeleteDevice(w, r, client)
		default:
			handler.HTTPJsonMsg(w, APIMethodNotAllowed, http.StatusMethodNotAllowed)
		}
	})

	http.HandleFunc("/session", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case MethodGet:
			handler.HandleGetSession(w, r)
		default:
			handler.HTTPJsonMsg(w, APIMethodNotAllowed, http.StatusMethodNotAllowed)
		}
	})

	http.HandleFunc("/refresh", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case MethodPut:
			handler.HandlePutRefreshToken(w, r)
		default:
			handler.HTTPJsonMsg(w, APIMethodNotAllowed, http.StatusMethodNotAllowed)
		}
	})

	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatal(err)
	}
}
