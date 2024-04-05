package main

import (
	"log"
	"net/http"

	"github.com/DerRomtester/testproject-golang-webapi-1/handler"
	"github.com/DerRomtester/testproject-golang-webapi-1/model"

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
	APIMethodNotAllowed model.APIError
	err error
)

func main() {
	APIMethodNotAllowed.Err = "method not allowed"

	http.HandleFunc("/auth", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case MethodPut:
			handler.HandlePutLogout(w, r, client)
		case MethodPost:
			client, err = handler.HandlePostLogin(w, r)
			if err != nil {
				log.Fatal(err)
			}
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

	err = http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatal(err)
	}
}
