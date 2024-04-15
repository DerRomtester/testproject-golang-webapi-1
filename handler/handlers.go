package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/DerRomtester/testproject-golang-webapi-1/database"
	"github.com/DerRomtester/testproject-golang-webapi-1/model"

	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

var (
	ErrorMsg                model.APIError
	ErrWrongStructure       = "structure of request is wrong"
	ErrAlreadyAuthenticated = "client already authenticated"
	ErrNotAuthenticated     = "not authenticated"
	ErrSessionNotExist      = "session does not exist"
	ErrSessionExpired       = "session expired"
	ErrNoCookie             = "no session cookie"
	ErrNoDeviceID           = "deviceID needs to be specified"
	ErrDatabase             = "db error"
)

type Authorization interface {
	CheckAuth(r *http.Request) (model.APIError, error)
	CheckAuthValidJson(r *http.Request) (model.UserCredentials, model.APIError, error)
}

func CheckAuthValidJson(r *http.Request) (model.UserCredentials, model.APIError, error) {
	var creds model.UserCredentials
	var msg model.APIError
	err := json.NewDecoder(r.Body).Decode(&creds)
	if err != nil {
		msg = model.APIError{
			Err: ErrWrongStructure,
		}
		return creds, msg, err
	}
	return creds, msg, nil
}

func HandleCreateUser(w http.ResponseWriter, r *http.Request, client *mongo.Client) error {
	creds, msg, err := CheckAuthValidJson(r)
	if err != nil {
		HTTPJsonMsg(w, msg, http.StatusBadRequest)
		return err
	}

	if err := database.ClientStatusDB(client); err != nil {
		ErrorMsg.Err = ErrNotAuthenticated
		HTTPJsonMsg(w, ErrorMsg, http.StatusUnauthorized)
		return err
	}

	if err := database.CreateUserDB(creds, client); err != nil {
		ErrorMsg.Err = err.Error()
		HTTPJsonMsg(w, ErrorMsg, http.StatusInternalServerError)
		return nil
	}
	return nil
}

func CheckUserPassword(u model.UserCredentials, c *mongo.Client) error {
	db, err := database.GetUserDB(u.Username, c)
	if err != nil {
		return err
	}

	if u.Username == db.Username && u.Password == db.Password {
		return nil
	}
	return errors.New("username or password do not match")
}

func HandlePostLogin(w http.ResponseWriter, r *http.Request, c *mongo.Client) error {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	var (
		err error
		msg model.APIError
	)

	_, err = CheckAuth(r, c)
	if err == nil {
		msg := model.APIError{
			Err: ErrAlreadyAuthenticated,
		}
		HTTPJsonMsg(w, msg, http.StatusAlreadyReported)
		return errors.New(ErrAlreadyAuthenticated)
	}

	var creds model.UserCredentials
	// Get the JSON body and decode into UserCredentials
	creds, msg, err = CheckAuthValidJson(r)
	if err != nil {
		HTTPJsonMsg(w, msg, http.StatusBadRequest)
		return err
	}

	err = CheckUserPassword(creds, c)
	if err != nil {
		ErrorMsg.Err = ErrNotAuthenticated
		HTTPJsonMsg(w, ErrorMsg, http.StatusUnauthorized)
		return errors.New(ErrNotAuthenticated)
	}

	sessionToken := uuid.NewString()
	expiresAt := time.Now().Add(120 * time.Second)

	session := model.UserSession{
		Username: creds.Username,
		Expiry:   expiresAt,
		Token:    sessionToken,
	}

	err = database.CreateSessionDB(session, c)
	if err != nil {
		ErrorMsg.Err = ErrDatabase
		HTTPJsonMsg(w, ErrorMsg, http.StatusInternalServerError)
		return errors.New(ErrDatabase)
	}

	http.SetCookie(w, &http.Cookie{
		Name:    "session_token",
		Value:   sessionToken,
		Expires: expiresAt,
	})

	return nil
}

func CheckAuth(r *http.Request, client *mongo.Client) (model.APIError, error) {
	var ErrorMsg model.APIError
	c, err := r.Cookie("session_token")
	if err != nil {
		if err == http.ErrNoCookie {
			ErrorMsg.Err = ErrNoCookie
			return ErrorMsg, err
		}
		return ErrorMsg, err
	}
	sessionToken := c.Value
	userSession, err := database.GetTokenDB(sessionToken, client)
	if err != nil {
		ErrorMsg.Err = ErrSessionNotExist
		return ErrorMsg, errors.New(ErrSessionNotExist)
	}

	if userSession.IsExpired() {
		err = database.DeleteTokenDB(sessionToken, client)
		if err != nil {
			ErrorMsg.Err = ErrDatabase
			return ErrorMsg, errors.New(ErrDatabase)
		}
		ErrorMsg.Err = ErrSessionExpired
		return ErrorMsg, errors.New(ErrSessionExpired)
	}
	return ErrorMsg, nil
}

func HandleGetSession(w http.ResponseWriter, r *http.Request, client *mongo.Client) error {
	c, err := r.Cookie("session_token")
	if err != nil {
		if err == http.ErrNoCookie {
			ErrorMsg.Err = ErrNoCookie
			HTTPJsonMsg(w, ErrorMsg, http.StatusUnauthorized)
			return err
		}
		w.WriteHeader(http.StatusBadRequest)
		return err
	}
	sessionToken := c.Value
	userSession, err := database.GetTokenDB(sessionToken, client)
	if err != nil {
		ErrorMsg.Err = ErrSessionNotExist
		HTTPJsonMsg(w, ErrorMsg, http.StatusUnauthorized)
		return err
	}

	if userSession.IsExpired() {
		database.DeleteTokenDB(sessionToken, client)
		ErrorMsg.Err = ErrSessionExpired
		HTTPJsonMsg(w, ErrorMsg, http.StatusUnauthorized)
		return err
	}

	HTTPJsonMsg(w, userSession, http.StatusOK)
	return nil
}

func HandlePutRefreshToken(w http.ResponseWriter, r *http.Request, client *mongo.Client) error {
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

	userSession, err := database.GetTokenDB(sessionToken, client)
	if err != nil {
		ErrorMsg.Err = ErrSessionNotExist
		HTTPJsonMsg(w, ErrorMsg, http.StatusUnauthorized)
		return err
	}

	if userSession.IsExpired() {
		database.DeleteTokenDB(sessionToken, client)
		ErrorMsg.Err = ErrSessionExpired
		HTTPJsonMsg(w, ErrorMsg, http.StatusUnauthorized)
		return err
	}

	newSession := userSession.RenewSession(120)
	err = database.CreateSessionDB(newSession, client)
	if err != nil {
		return err
	}

	database.DeleteTokenDB(sessionToken, client)

	http.SetCookie(w, &http.Cookie{
		Name:    "session_token",
		Value:   newSession.Token,
		Expires: time.Now().Add(120 * time.Second),
	})
	return nil
}

func HandlePutLogout(w http.ResponseWriter, r *http.Request, client *mongo.Client) error {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	_, err := CheckAuth(r, client)
	if err != nil {
		ErrorMsg.Err = ErrNotAuthenticated
		HTTPJsonMsg(w, ErrorMsg, http.StatusBadRequest)
		return errors.New(ErrorMsg.Err)
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

	err = database.DeleteTokenDB(sessionToken, client)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return err
	}

	http.SetCookie(w, &http.Cookie{
		Name:   "session_token",
		Value:  "",
		MaxAge: -1,
	})

	err = database.ClientStatusDB(client)
	if err != nil {
		return err
	}

	return nil
}

func HandleGetDevices(w http.ResponseWriter, r *http.Request, client *mongo.Client) error {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	msg, err := CheckAuth(r, client)
	if err != nil {
		HTTPJsonMsg(w, msg, http.StatusUnauthorized)
		return err
	}

	err = database.ClientStatusDB(client)
	if err != nil {
		ErrorMsg.Err = ErrNotAuthenticated
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

	msg, err := CheckAuth(r, client)
	if err != nil {
		HTTPJsonMsg(w, msg, http.StatusUnauthorized)
		return nil
	}

	var devices model.Root
	err = database.ClientStatusDB(client)
	if err != nil {
		ErrorMsg.Err = ErrNotAuthenticated
		HTTPJsonMsg(w, ErrorMsg, http.StatusUnauthorized)
		return err
	}

	if id == "" {
		ErrorMsg.Err = ErrNoDeviceID
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

	msg, err := CheckAuth(r, client)
	if err != nil {
		HTTPJsonMsg(w, msg, http.StatusUnauthorized)
		return nil
	}

	err = database.ClientStatusDB(client)
	if err != nil {
		ErrorMsg.Err = ErrNotAuthenticated
		HTTPJsonMsg(w, ErrorMsg, http.StatusUnauthorized)
		return err
	}

	if id == "" {
		ErrorMsg.Err = ErrNoDeviceID
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

	msg, err := CheckAuth(r, client)
	if err != nil {
		HTTPJsonMsg(w, msg, http.StatusUnauthorized)
		return err
	}

	err = database.ClientStatusDB(client)
	if err != nil {
		ErrorMsg.Err = ErrNotAuthenticated
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

	msg, err := CheckAuth(r, client)
	if err != nil {
		HTTPJsonMsg(w, msg, http.StatusUnauthorized)
		return nil
	}

	err = database.ClientStatusDB(client)
	if err != nil {
		ErrorMsg.Err = ErrNotAuthenticated
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
