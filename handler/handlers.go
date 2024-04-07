package handler

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/DerRomtester/testproject-golang-webapi-1/database"
	"github.com/DerRomtester/testproject-golang-webapi-1/model"

	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

var ErrorMsg model.APIError
var users = map[string]string{
	"user1": "password1",
	"user2": "password2",
}

type session struct {
	Username string    `json:"username"`
	Expiry   time.Time `json:"expiry"`
}

func (s session) isExpired() bool {
	return s.Expiry.Before(time.Now())
}

var sessions = map[string]session{}

func HandlePostLogin(w http.ResponseWriter, r *http.Request) (*mongo.Client, error) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	var creds model.Credentials
	// Get the JSON body and decode into credentials
	err := json.NewDecoder(r.Body).Decode(&creds)
	if err != nil {
		ErrorMsg.Err = "structure of request is wrong"
		HTTPJsonMsg(w, ErrorMsg, http.StatusBadRequest)
		return nil, err
	}

	expectedPassword, ok := users[creds.Username]

	if !ok || expectedPassword != creds.Password {
		ErrorMsg.Err = "not authorized"
		HTTPJsonMsg(w, ErrorMsg, http.StatusUnauthorized)
		return nil, errors.New("Not authorized")
	}

	sessionToken := uuid.NewString()
	expiresAt := time.Now().Add(120 * time.Second)

	sessions[sessionToken] = session{
		Username: creds.Username,
		Expiry:   expiresAt,
	}

	http.SetCookie(w, &http.Cookie{
		Name:    "session_token",
		Value:   sessionToken,
		Expires: expiresAt,
	})

	client, err := database.ConnectDB("mongodb://localhost:27017")
	if err != nil {
		ErrorMsg.Err = "Error connecting to Database"
		HTTPJsonMsg(w, ErrorMsg, http.StatusInternalServerError)
		return nil, err
	}
	return client, nil
}

func CheckAuth(r *http.Request) (model.APIError, error) {
	var ErrorMsg model.APIError
	c, err := r.Cookie("session_token")
	if err != nil {
		if err == http.ErrNoCookie {
			ErrorMsg.Err = "session does not exist"
			return ErrorMsg, err
		}
		return ErrorMsg, err
	}
	sessionToken := c.Value
	userSession, exists := sessions[sessionToken]
	if !exists {
		ErrorMsg.Err = "session does not exist"
		return ErrorMsg, errors.New("Session does not exist")
	}
	if userSession.isExpired() {
		delete(sessions, sessionToken)
		ErrorMsg.Err = "session is expired"
		return ErrorMsg, errors.New("Session is expired")
	}
	return ErrorMsg, nil
}

func HandleGetSession(w http.ResponseWriter, r *http.Request) error {
	c, err := r.Cookie("session_token")
	if err != nil {
		if err == http.ErrNoCookie {
			ErrorMsg.Err = "no session cookie found"
			HTTPJsonMsg(w, ErrorMsg, http.StatusUnauthorized)
			return err
		}
		w.WriteHeader(http.StatusBadRequest)
		return err
	}
	sessionToken := c.Value
	userSession, exists := sessions[sessionToken]
	if !exists {
		ErrorMsg.Err = "session does not exist"
		HTTPJsonMsg(w, ErrorMsg, http.StatusUnauthorized)
		return err
	}
	if userSession.isExpired() {
		delete(sessions, sessionToken)
		ErrorMsg.Err = "session is expired"
		HTTPJsonMsg(w, ErrorMsg, http.StatusUnauthorized)
		return err
	}
	HTTPJsonMsg(w, userSession, http.StatusOK)
	return nil
}

func HandlePutRefreshToken(w http.ResponseWriter, r *http.Request) error {
	c, err := r.Cookie("session_token")
	if err != nil {
		if err == http.ErrNoCookie {
			w.WriteHeader(http.StatusUnauthorized)
			return err
		}
		w.WriteHeader(http.StatusBadRequest)
		return err
	}
	sessionToken := c.Value

	userSession, exists := sessions[sessionToken]
	if !exists {
		w.WriteHeader(http.StatusUnauthorized)
		return errors.New("Session does not exist")
	}
	if userSession.isExpired() {
		delete(sessions, sessionToken)
		w.WriteHeader(http.StatusUnauthorized)
		return errors.New("Session is expired")
	}

	newSessionToken := uuid.NewString()
	expiresAt := time.Now().Add(120 * time.Second)

	sessions[newSessionToken] = session{
		Username: userSession.Username,
		Expiry:   expiresAt,
	}

	delete(sessions, sessionToken)

	http.SetCookie(w, &http.Cookie{
		Name:    "session_token",
		Value:   newSessionToken,
		Expires: time.Now().Add(120 * time.Second),
	})
	return nil
}

func HandlePutLogout(w http.ResponseWriter, r *http.Request, client *mongo.Client) error {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	var (
		c, err       = r.Cookie("session_token")
		errs         []error
		sessionToken = c.Value
	)

	if err != nil {
		if err == http.ErrNoCookie {
			w.WriteHeader(http.StatusUnauthorized)
			return nil
		}
		w.WriteHeader(http.StatusBadRequest)
		return err
	}

	delete(sessions, sessionToken)

	http.SetCookie(w, &http.Cookie{
		Name:   "session_token",
		Value:  "",
		MaxAge: -1,
	})

	if err = database.ClientStatusDB(client); err != nil {
		errs = append(errs, err)
	}

	if err := client.Disconnect(context.TODO()); err != nil {
		errs = append(errs, err)
	}

	if len(errs) > 0 {
		w.WriteHeader(http.StatusInternalServerError)
		return fmt.Errorf("logout errors: %w", errs) // Combine errors
	}

	return nil
}

func HandleGetDevices(w http.ResponseWriter, r *http.Request, client *mongo.Client) error {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	msg, err := CheckAuth(r)
	if err != nil {
		HTTPJsonMsg(w, msg, http.StatusUnauthorized)
		return err
	}

	err = database.ClientStatusDB(client)
	if err != nil {
		ErrorMsg.Err = "client not authenticated"
		HTTPJsonMsg(w, ErrorMsg, http.StatusUnauthorized)
		return err
	}

	var devices model.Root
	devices, err = database.GetDeviceDB(bson.D{{}}, client)
	if err != nil {
		ErrorMsg.Err = err.Error()
		HTTPJsonMsg(w, ErrorMsg, http.StatusInternalServerError)
		return err
	}
	alldevices, err := json.Marshal(devices)
	if err != nil {
		ErrorMsg.Err = err.Error()
		HTTPJsonMsg(w, ErrorMsg, http.StatusInternalServerError)
		return err
	}
	w.Write(alldevices)
	return nil
}

func HandleGetDeviceByID(w http.ResponseWriter, r *http.Request, client *mongo.Client, id string) error {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	msg, err := CheckAuth(r)
	if err != nil {
		HTTPJsonMsg(w, msg, http.StatusUnauthorized)
		return nil
	}

	var devices model.Root
	err = database.ClientStatusDB(client)
	if err != nil {
		ErrorMsg.Err = "client not authenticated"
		HTTPJsonMsg(w, ErrorMsg, http.StatusUnauthorized)
		return err
	}

	if id == "" {
		ErrorMsg.Err = "device id must be specified"
		HTTPJsonMsg(w, ErrorMsg, http.StatusBadRequest)
		return nil
	}

	devices, err = database.GetDeviceDB(primitive.D{{Key: "_id", Value: id}}, client)
	if err != nil {
		ErrorMsg.Err = err.Error()
		HTTPJsonMsg(w, ErrorMsg, http.StatusInternalServerError)
		return err
	}
	singledevice, err := json.Marshal(devices)
	if err != nil {
		ErrorMsg.Err = err.Error()
		HTTPJsonMsg(w, ErrorMsg, http.StatusInternalServerError)
		return err
	}
	w.Write(singledevice)
	return nil
}

func HandleDeleteDevice(w http.ResponseWriter, r *http.Request, client *mongo.Client, id string) error {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	msg, err := CheckAuth(r)
	if err != nil {
		HTTPJsonMsg(w, msg, http.StatusUnauthorized)
		return nil
	}

	err = database.ClientStatusDB(client)
	if err != nil {
		ErrorMsg.Err = "client not authenticated"
		HTTPJsonMsg(w, ErrorMsg, http.StatusUnauthorized)
		return err
	}

	if id == "" {
		ErrorMsg.Err = "device id must be specified"
		HTTPJsonMsg(w, ErrorMsg, http.StatusBadRequest)
		return nil
	}

	err = database.DeleteDeviceDB(primitive.D{{Key: "_id", Value: id}}, client)
	if err != nil {
		ErrorMsg.Err = err.Error()
		HTTPJsonMsg(w, ErrorMsg, http.StatusInternalServerError)
		return nil
	}
	return nil
}

func HandleDeleteDevices(w http.ResponseWriter, r *http.Request, client *mongo.Client) error {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	msg, err := CheckAuth(r)
	if err != nil {
		HTTPJsonMsg(w, msg, http.StatusUnauthorized)
		return err
	}

	err = database.ClientStatusDB(client)
	if err != nil {
		ErrorMsg.Err = "client not authenticated"
		HTTPJsonMsg(w, ErrorMsg, http.StatusUnauthorized)
		return err
	}

	err = database.DeleteDevicesDB(bson.D{{}}, client)
	if err != nil {
		ErrorMsg.Err = err.Error()
		HTTPJsonMsg(w, ErrorMsg, http.StatusInternalServerError)
		return err
	}
	return nil
}

func HandlePostDevices(w http.ResponseWriter, r *http.Request, client *mongo.Client) error {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	var devices model.Root

	msg, err := CheckAuth(r)
	if err != nil {
		HTTPJsonMsg(w, msg, http.StatusUnauthorized)
		return nil
	}

	err = database.ClientStatusDB(client)
	if err != nil {
		ErrorMsg.Err = "client not authenticated"
		HTTPJsonMsg(w, ErrorMsg, http.StatusUnauthorized)
		return err
	}

	err = json.NewDecoder(r.Body).Decode(&devices)

	if err != nil {
		ErrorMsg.Err = err.Error()
		HTTPJsonMsg(w, ErrorMsg, http.StatusBadRequest)
		return nil
	}

	if err := database.WriteDevicesDB(devices, client); err != nil {
		ErrorMsg.Err = err.Error()
		HTTPJsonMsg(w, ErrorMsg, http.StatusInternalServerError)
		return nil
	}
	return nil
}

func HTTPJsonMsg(w http.ResponseWriter, err interface{}, code int) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(err)
}
