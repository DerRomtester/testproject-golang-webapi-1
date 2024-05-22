package handler

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/DerRomtester/testproject-golang-webapi-1/internal/database"
	"github.com/DerRomtester/testproject-golang-webapi-1/internal/helper"
	"github.com/DerRomtester/testproject-golang-webapi-1/internal/session"
	"github.com/DerRomtester/testproject-golang-webapi-1/model"

	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

var (
	ErrWrongStructure       = APIError{Code: 400, Message: "structure of request is wrong"}
	ErrAlreadyAuthenticated = APIError{Code: 409, Message: "client already authenticated"}
	ErrNotAuthenticated     = APIError{Code: 401, Message: "not authenticated"}
	ErrSessionNotExist      = APIError{Code: 401, Message: "session does not exist"}
	ErrSessionExpired       = APIError{Code: 401, Message: "session expired"}
	ErrNoCookie             = APIError{Code: 401, Message: "no session cookie"}
	ErrNoDeviceID           = APIError{Code: 400, Message: "deviceID needs to be specified"}
	ErrDatabase             = APIError{Code: 500, Message: "db error"}
	ErrHashingPW            = APIError{Code: 401, Message: "failed to hash password"}
)

type APIError struct {
	Code    int    `json:"code"`
	Message string `json:"error"`
}

type ServerConfig struct {
	Domain string
	Port   string
}

type Server struct {
	Server *http.ServeMux
}

func (e *APIError) CustomError() error {
	return fmt.Errorf("%d %s", ErrNotAuthenticated.Code, ErrNotAuthenticated.Message)
}

func (s *ServerConfig) GetHost() *string {
	return &s.Domain
}

func (s *ServerConfig) GetDomain() *string {
	return &s.Port
}

func (s *ServerConfig) Run(mg database.DBClient) {
	mux := http.NewServeMux()
	mux.HandleFunc("PUT /v1/auth", func(w http.ResponseWriter, r *http.Request) {
		HandlePutLogout(w, r, mg)
	})

	mux.HandleFunc("POST /v1/auth", func(w http.ResponseWriter, r *http.Request) {
		HandlePostLogin(w, r, mg)
	})

	mux.HandleFunc("GET /v1/devices", func(w http.ResponseWriter, r *http.Request) {
		HandleGetDevices(w, r, mg)
	})

	mux.HandleFunc("POST /v1/devices", func(w http.ResponseWriter, r *http.Request) {
		HandlePostDevices(w, r, mg)
	})

	mux.HandleFunc("DELETE /v1/devices", func(w http.ResponseWriter, r *http.Request) {
		HandleDeleteDevices(w, r, mg)
	})

	mux.HandleFunc("GET /v1/device/{id}", func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		HandleGetDeviceByID(w, r, mg, id)
	})

	mux.HandleFunc("DELETE /v1/device/{id}", func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		HandleDeleteDevice(w, r, mg, id)
	})

	mux.HandleFunc("GET /v1/session", func(w http.ResponseWriter, r *http.Request) {
		HandleGetSession(w, r, mg)
	})

	mux.HandleFunc("PUT /v1/refresh", func(w http.ResponseWriter, r *http.Request) {
		HandlePutRefreshToken(w, r, mg)
	})

	mux.HandleFunc("POST /v1/user", func(w http.ResponseWriter, r *http.Request) {
		HandleCreateUser(w, r, mg)
	})

	log.Println("starting server on host: ", s.Domain, " Port: ", s.Port)
	http.ListenAndServe(s.Domain+s.Port, mux)
}

func CheckAuthValidJson(r *http.Request) (model.UserCredentials, APIError, error) {
	var creds model.UserCredentials
	var msg APIError
	err := json.NewDecoder(r.Body).Decode(&creds)

	if err != nil {
		return creds, ErrWrongStructure, err
	}

	return creds, msg, nil
}

func HandleCreateUser(w http.ResponseWriter, r *http.Request, mg database.DBClient) error {
	creds, msg, err := CheckAuthValidJson(r)
	if err != nil {
		HTTPJsonMsg(w, msg, http.StatusBadRequest)
		return err
	}

	if err := mg.ClientStatusDB(); err != nil {
		HTTPJsonMsg(w, ErrNotAuthenticated, http.StatusUnauthorized)
		return err
	}

	if err := mg.CreateUserDB(creds); err != nil {
		HTTPJsonMsg(w, err.Error(), http.StatusInternalServerError)
		return err
	}
	return nil
}

func CheckUserPassword(u model.UserCredentials, mg database.DBClient) error {
	db, err := mg.GetUserDB(u.Username)
	if err != nil {
		return err
	}

	err = helper.CheckPassword(u.Password, db.Password)

	if err != nil {
		return err
	}

	return nil
}

func HandlePostLogin(w http.ResponseWriter, r *http.Request, mg database.DBClient) error {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	_, err := CheckAuth(r, mg)
	if err == nil {
		HTTPJsonMsg(w, ErrAlreadyAuthenticated, http.StatusAlreadyReported)
		return ErrAlreadyAuthenticated.CustomError()
	}

	var creds model.UserCredentials
	// Get the JSON body and decode into UserCredentials
	creds, msg, err := CheckAuthValidJson(r)
	if err != nil {
		HTTPJsonMsg(w, msg, http.StatusBadRequest)
		return err
	}

	err = CheckUserPassword(creds, mg)
	if err != nil {
		HTTPJsonMsg(w, ErrNotAuthenticated, http.StatusUnauthorized)
		return ErrNotAuthenticated.CustomError()
	}

	sessionToken := uuid.NewString()
	expiresAt := time.Now().Add(120 * time.Second)

	session := session.UserSession{
		Username: creds.Username,
		Expiry:   expiresAt,
		Token:    sessionToken,
	}

	err = mg.CreateSessionDB(session)
	if err != nil {
		HTTPJsonMsg(w, ErrDatabase, http.StatusInternalServerError)
		return ErrDatabase.CustomError()
	}

	http.SetCookie(w, &http.Cookie{
		Name:    "session_token",
		Value:   sessionToken,
		Expires: expiresAt,
	})

	return nil
}

func CheckAuth(r *http.Request, mg database.DBClient) (APIError, error) {
	c, err := r.Cookie("session_token")
	if err != nil {
		if err == http.ErrNoCookie {
			return ErrNoCookie, err
		}
		return ErrNoCookie, err
	}
	sessionToken := c.Value
	existingToken, err := mg.GetTokenDB(sessionToken)
	if err != nil {
		return ErrSessionNotExist, ErrSessionNotExist.CustomError()
	}

	if existingToken.IsExpired() {
		err = mg.DeleteTokenDB(*existingToken.GetToken())
		if err != nil {
			return ErrDatabase, ErrDatabase.CustomError()
		}
		return ErrSessionExpired, ErrSessionNotExist.CustomError()
	}
	return APIError{}, nil
}

func HandleGetSession(w http.ResponseWriter, r *http.Request, mg database.DBClient) error {
	c, err := r.Cookie("session_token")
	if err != nil {
		if err == http.ErrNoCookie {
			HTTPJsonMsg(w, ErrNoCookie, http.StatusUnauthorized)
			return err
		}
		w.WriteHeader(http.StatusBadRequest)
		return err
	}
	sessionToken := c.Value
	userSession, err := mg.GetTokenDB(sessionToken)
	if err != nil {
		HTTPJsonMsg(w, ErrSessionNotExist, http.StatusUnauthorized)
		return err
	}

	if userSession.IsExpired() {
		mg.DeleteTokenDB(sessionToken)
		HTTPJsonMsg(w, ErrSessionExpired, http.StatusUnauthorized)
		return err
	}

	HTTPJsonMsg(w, userSession, http.StatusOK)
	return nil
}

func HandlePutRefreshToken(w http.ResponseWriter, r *http.Request, mg database.DBClient) error {
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

	userSession, err := mg.GetTokenDB(sessionToken)
	if err != nil {
		HTTPJsonMsg(w, ErrSessionNotExist, http.StatusUnauthorized)
		return err
	}

	if userSession.IsExpired() {
		mg.DeleteTokenDB(sessionToken)
		HTTPJsonMsg(w, ErrSessionExpired, http.StatusUnauthorized)
		return err
	}

	newSession := userSession.RenewSession(120)
	err = mg.CreateSessionDB(newSession)
	if err != nil {
		return err
	}

	mg.DeleteTokenDB(sessionToken)

	http.SetCookie(w, &http.Cookie{
		Name:    "session_token",
		Value:   newSession.Token,
		Expires: time.Now().Add(120 * time.Second),
	})
	return nil
}

func HandlePutLogout(w http.ResponseWriter, r *http.Request, mg database.DBClient) error {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	_, err := CheckAuth(r, mg)
	if err != nil {
		HTTPJsonMsg(w, ErrNotAuthenticated, http.StatusBadRequest)
		return ErrNotAuthenticated.CustomError()
	}

	c, err := r.Cookie("session_token")
	sessionToken := c.Value

	if err != nil {
		if err == http.ErrNoCookie {
			w.WriteHeader(http.StatusUnauthorized)
			return nil
		}
		w.WriteHeader(http.StatusBadRequest)
		return err
	}

	err = mg.DeleteTokenDB(sessionToken)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return err
	}

	http.SetCookie(w, &http.Cookie{
		Name:   "session_token",
		Value:  "",
		MaxAge: -1,
	})

	err = mg.ClientStatusDB()
	if err != nil {
		return err
	}

	return nil
}

func HandleGetDevices(w http.ResponseWriter, r *http.Request, mg database.DBClient) error {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	msg, err := CheckAuth(r, mg)
	if err != nil {
		HTTPJsonMsg(w, msg, http.StatusUnauthorized)
		return err
	}

	err = mg.ClientStatusDB()
	if err != nil {
		HTTPJsonMsg(w, ErrNotAuthenticated, http.StatusUnauthorized)
		return ErrNotAuthenticated.CustomError()
	}

	var devices model.Devices
	devices, err = mg.GetDeviceDB(bson.D{{}})
	if err != nil {
		HTTPJsonMsg(w, err, http.StatusInternalServerError)
		return err
	}
	alldevices, err := json.Marshal(devices)
	if err != nil {
		HTTPJsonMsg(w, err, http.StatusInternalServerError)
		return err
	}
	w.Write(alldevices)
	return nil
}

func HandleGetDeviceByID(w http.ResponseWriter, r *http.Request, mg database.DBClient, id string) error {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	msg, err := CheckAuth(r, mg)
	if err != nil {
		HTTPJsonMsg(w, msg, http.StatusUnauthorized)
		return nil
	}

	var devices model.Devices
	err = mg.ClientStatusDB()
	if err != nil {
		HTTPJsonMsg(w, ErrNotAuthenticated, http.StatusUnauthorized)
		return ErrNotAuthenticated.CustomError()
	}

	if id == "" {
		HTTPJsonMsg(w, ErrNoDeviceID, http.StatusBadRequest)
		return ErrNoDeviceID.CustomError()
	}

	devices, err = mg.GetDeviceDB(primitive.D{{Key: "_id", Value: id}})
	if err != nil {
		HTTPJsonMsg(w, err, http.StatusInternalServerError)
		return err
	}
	singledevice, err := json.Marshal(devices)
	if err != nil {
		HTTPJsonMsg(w, err, http.StatusInternalServerError)
		return err
	}
	w.Write(singledevice)
	return nil
}

func HandleDeleteDevice(w http.ResponseWriter, r *http.Request, mg database.DBClient, id string) error {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	msg, err := CheckAuth(r, mg)
	if err != nil {
		HTTPJsonMsg(w, msg, http.StatusUnauthorized)
		return nil
	}

	err = mg.ClientStatusDB()
	if err != nil {
		HTTPJsonMsg(w, ErrNotAuthenticated, http.StatusUnauthorized)
		return ErrNotAuthenticated.CustomError()
	}

	if id == "" {
		HTTPJsonMsg(w, ErrNoDeviceID, http.StatusBadRequest)
		return ErrNoDeviceID.CustomError()
	}

	err = mg.DeleteDeviceDB(primitive.D{{Key: "_id", Value: id}}, false)
	if err != nil {
		HTTPJsonMsg(w, err, http.StatusInternalServerError)
		return err
	}
	return nil
}

func HandleDeleteDevices(w http.ResponseWriter, r *http.Request, mg database.DBClient) error {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	msg, err := CheckAuth(r, mg)
	if err != nil {
		HTTPJsonMsg(w, msg, http.StatusUnauthorized)
		return err
	}

	err = mg.ClientStatusDB()
	if err != nil {
		HTTPJsonMsg(w, ErrNotAuthenticated, http.StatusUnauthorized)
		return ErrNotAuthenticated.CustomError()
	}

	err = mg.DeleteDeviceDB(bson.D{{}}, true)
	if err != nil {
		HTTPJsonMsg(w, err, http.StatusInternalServerError)
		return err
	}
	return nil
}

func HandlePostDevices(w http.ResponseWriter, r *http.Request, mg database.DBClient) error {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	var devices model.Devices

	msg, err := CheckAuth(r, mg)
	if err != nil {
		HTTPJsonMsg(w, msg, http.StatusUnauthorized)
		return nil
	}

	err = mg.ClientStatusDB()
	if err != nil {
		HTTPJsonMsg(w, err, http.StatusUnauthorized)
		return err
	}

	err = json.NewDecoder(r.Body).Decode(&devices)

	if err != nil {
		HTTPJsonMsg(w, err, http.StatusBadRequest)
		return err
	}

	if err := mg.WriteDevicesDB(devices); err != nil {
		HTTPJsonMsg(w, err, http.StatusInternalServerError)
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
