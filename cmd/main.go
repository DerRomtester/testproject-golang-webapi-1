package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Root struct {
	Devices []Device `json:"devices"`
}

type Device struct {
	ID                              string `bson:"_id,omitempty"`
	Name                            string `json:"name"`
	DeviceTypeID                    string `json:"deviceTypeId"`
	Failsafe                        bool   `json:"failsafe"`
	TempMin                         int    `json:"tempMin"`
	TempMax                         int    `json:"tempMax"`
	InstallationPosition            string `json:"installationPosition"`
	InsertInto19InchCabinet         bool   `json:"insertInto19InchCabinet"`
	MotionEnable                    bool   `json:"motionEnable"`
	SiplusCatalog                   bool   `json:"siplusCatalog"`
	SimaticCatalog                  bool   `json:"simaticCatalog"`
	RotationAxisNumber              int    `json:"rotationAxisNumber"`
	PositionAxisNumber              int    `json:"positionAxisNumber"`
	AdvancedEnvironmentalConditions bool   `json:"advancedEnvironmentalConditions,omitempty"`
	TerminalElement                 bool   `json:"terminalElement,omitempty"`
}

type Credentials struct {
	Password string `json:"password"`
	Username string `json:"username"`
}

type APIError struct {
	Err string `json:"error"`
}

var sessions = map[string]session{}

type session struct {
	Username string    `json:"username"`
	Expiry   time.Time `json:"expiry"`
}

func (s session) isExpired() bool {
	return s.Expiry.Before(time.Now())
}

const (
	MethodGet    = "GET"
	MethodPost   = "POST"
	MethodPut    = "PUT"
	MethodDelete = "DELETE"
)

var users = map[string]string{
	"user1": "password1",
	"user2": "password2",
}

var (
	client   *mongo.Client
	ErrorMsg APIError
)

func ConnectDB(mongoURI string) (*mongo.Client, error) {
	client, err := mongo.Connect(context.TODO(), options.Client().ApplyURI(mongoURI))
	if err != nil {
		fmt.Println("Error connecting to database")
		return nil, err
	}

	err = client.Ping(context.Background(), nil)
	if err != nil {
		fmt.Println("Error pinging database")
		return nil, err
	}

	fmt.Println("Connected to MongoDb")
	return client, nil
}

func ClientStatusDB(client *mongo.Client) error {
	if client == nil {
		return errors.New("error: no client connection to database")
	}
	return nil
}

func GetDeviceDB(filter bson.D, c *mongo.Client) (Root, error) {
	var devices Root
	err := ClientStatusDB(client)
	if err != nil {
		return devices, err
	}

	collection := c.Database("devices-db").Collection("Devices")
	cursor, err := collection.Find(context.Background(), filter)
	if err != nil {
		return devices, err
	}
	defer cursor.Close(context.Background())

	for cursor.Next(context.Background()) {
		var device Device
		if err := cursor.Decode(&device); err != nil {
			return devices, err
		}
		devices.Devices = append(devices.Devices, device)
	}

	if err := cursor.Err(); err != nil {
		return devices, err
	}

	return devices, nil
}

func DeleteDeviceDB(filter bson.D, c *mongo.Client) error {
	err := ClientStatusDB(client)
	if err != nil {
		return err
	}

	collection := c.Database("devices-db").Collection("Devices")
	if filter == nil {
		err := errors.New("device id must be specified")
		return err
	}

	_, err = collection.DeleteMany(context.TODO(), filter)
	if err != nil {
		return err
	}
	return nil
}

func DeleteDevicesDB(filter bson.D, c *mongo.Client) error {
	err := ClientStatusDB(client)
	if err != nil {
		return err
	}

	collection := c.Database("devices-db").Collection("Devices")
	_, err = collection.DeleteMany(context.TODO(), filter)
	if err != nil {
		return err
	}
	return nil
}

func WriteDevicesDB(devices Root, c *mongo.Client) error {
	err := ClientStatusDB(client)
	if err != nil {
		return err
	}

	collection := c.Database("devices-db").Collection("Devices")
	for _, device := range devices.Devices {
		filter := bson.D{primitive.E{Key: "_id", Value: device.ID}}
		var existingDevice Device
		err := collection.FindOne(context.Background(), filter).Decode(&existingDevice)
		if err == nil {
			update := bson.D{primitive.E{Key: "$set", Value: bson.M{"name": device.Name}}}
			_, err := collection.UpdateOne(context.Background(), filter, update)
			if err != nil {
				return err
			}
		} else {
			_, err := collection.InsertOne(context.Background(), device)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func HandlePostLogin(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	var creds Credentials
	// Get the JSON body and decode into credentials
	err := json.NewDecoder(r.Body).Decode(&creds)
	if err != nil {
		ErrorMsg.Err = "structure of request is wrong"
		HTTPJsonMsg(w, ErrorMsg, http.StatusBadRequest)
		return
	}

	// Get the expected password from our in memory map
	expectedPassword, ok := users[creds.Username]

	// If a password exists for the given user
	// AND, if it is the same as the password we received, the we can move ahead
	// if NOT, then we return an "Unauthorized" status
	if !ok || expectedPassword != creds.Password {
		ErrorMsg.Err = "not authorized"
		HTTPJsonMsg(w, ErrorMsg, http.StatusUnauthorized)
		return
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

	client, err = ConnectDB("mongodb://localhost:27017")
	if err != nil {
		ErrorMsg.Err = "Error connecting to Database"
		HTTPJsonMsg(w, ErrorMsg, http.StatusInternalServerError)
		return
	}
}

func HandleGetSession(w http.ResponseWriter, r *http.Request) {
	c, err := r.Cookie("session_token")
	if err != nil {
		if err == http.ErrNoCookie {
			ErrorMsg.Err = "no session cookie found"
			HTTPJsonMsg(w, ErrorMsg, http.StatusUnauthorized)
			return
		}
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	sessionToken := c.Value
	userSession, exists := sessions[sessionToken]
	if !exists {
		ErrorMsg.Err = "session does not exist"
		HTTPJsonMsg(w, ErrorMsg, http.StatusUnauthorized)
		return
	}
	if userSession.isExpired() {
		delete(sessions, sessionToken)
		ErrorMsg.Err = "session is expired"
		HTTPJsonMsg(w, ErrorMsg, http.StatusUnauthorized)
		return
	}
	HTTPJsonMsg(w, userSession, http.StatusOK)
}

func HandlePutRefreshToken(w http.ResponseWriter, r *http.Request) {
	c, err := r.Cookie("session_token")
	if err != nil {
		if err == http.ErrNoCookie {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	sessionToken := c.Value

	userSession, exists := sessions[sessionToken]
	if !exists {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	if userSession.isExpired() {
		delete(sessions, sessionToken)
		w.WriteHeader(http.StatusUnauthorized)
		return
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
}

func HandlePutLogout(w http.ResponseWriter, r *http.Request, client *mongo.Client) error {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	c, err := r.Cookie("session_token")
	if err != nil {
		if err == http.ErrNoCookie {
			// If the cookie is not set, return an unauthorized status
			w.WriteHeader(http.StatusUnauthorized)
			return err
		}
		// For any other type of error, return a bad request status
		w.WriteHeader(http.StatusBadRequest)
		return err
	}
	sessionToken := c.Value

	// remove the users session from the session map
	delete(sessions, sessionToken)

	http.SetCookie(w, &http.Cookie{
		Name:   "session_token",
		Value:  "",
		MaxAge: -1,
	})

	err = ClientStatusDB(client)
	if err != nil {
		ErrorMsg.Err = "client not authenticated"
		HTTPJsonMsg(w, ErrorMsg, http.StatusUnauthorized)
		return err
	}

	client.Disconnect(context.TODO())
	return nil
}

func HandleGetDevices(w http.ResponseWriter, r *http.Request, client *mongo.Client) error {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	err := ClientStatusDB(client)
	if err != nil {
		ErrorMsg.Err = "client not authenticated"
		HTTPJsonMsg(w, ErrorMsg, http.StatusUnauthorized)
		return err
	}

	var devices Root
	devices, err = GetDeviceDB(bson.D{{}}, client)
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

func HandleGetDeviceByID(w http.ResponseWriter, r *http.Request, client *mongo.Client) error {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	var devices Root
	err := ClientStatusDB(client)
	if err != nil {
		ErrorMsg.Err = "client not authenticated"
		HTTPJsonMsg(w, ErrorMsg, http.StatusUnauthorized)
		return err
	}

	id := r.URL.Query().Get("id")
	if id == "" {
		ErrorMsg.Err = "device id must be specified"
		HTTPJsonMsg(w, ErrorMsg, http.StatusBadRequest)
		return nil
	}

	devices, err = GetDeviceDB(primitive.D{{Key: "_id", Value: id}}, client)
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

func HandleDeleteDevice(w http.ResponseWriter, r *http.Request, client *mongo.Client) error {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	err := ClientStatusDB(client)
	if err != nil {
		ErrorMsg.Err = "client not authenticated"
		HTTPJsonMsg(w, ErrorMsg, http.StatusUnauthorized)
		return err
	}
	id := r.URL.Query().Get("id")
	if id == "" {
		ErrorMsg.Err = "device id must be specified"
		HTTPJsonMsg(w, ErrorMsg, http.StatusBadRequest)
		return nil
	}

	err = DeleteDeviceDB(primitive.D{{Key: "_id", Value: id}}, client)
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

	err := ClientStatusDB(client)
	if err != nil {
		ErrorMsg.Err = "client not authenticated"
		HTTPJsonMsg(w, ErrorMsg, http.StatusUnauthorized)
		return err
	}

	err = DeleteDevicesDB(bson.D{{}}, client)
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
	var devices Root

	err := ClientStatusDB(client)
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

	if err := WriteDevicesDB(devices, client); err != nil {
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

func main() {
	var APIMethodNotAllowed APIError
	APIMethodNotAllowed.Err = "method not allowed"

	http.HandleFunc("/auth", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case MethodPut:
			HandlePutLogout(w, r, client)
		case MethodPost:
			HandlePostLogin(w, r)
		default:
			HTTPJsonMsg(w, APIMethodNotAllowed, http.StatusMethodNotAllowed)
		}

	})
	http.HandleFunc("/devices", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case MethodGet:
			HandleGetDevices(w, r, client)
		case MethodPost:
			HandlePostDevices(w, r, client)
		case MethodDelete:
			HandleDeleteDevices(w, r, client)
		default:
			HTTPJsonMsg(w, APIMethodNotAllowed, http.StatusMethodNotAllowed)
		}

	})
	http.HandleFunc("/device", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == MethodGet {
			HandleGetDeviceByID(w, r, client)
		} else if r.Method == MethodDelete {
			HandleDeleteDevice(w, r, client)
		} else {
			HTTPJsonMsg(w, APIMethodNotAllowed, http.StatusMethodNotAllowed)
		}

	})

	http.HandleFunc("/session", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case MethodGet:
			HandleGetSession(w, r)
		default:
			HTTPJsonMsg(w, APIMethodNotAllowed, http.StatusMethodNotAllowed)
		}

	})
	http.HandleFunc("/refresh", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case MethodPut:
			HandlePutRefreshToken(w, r)
		default:
			HTTPJsonMsg(w, APIMethodNotAllowed, http.StatusMethodNotAllowed)
		}
	})

	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatal(err)
	}
}
