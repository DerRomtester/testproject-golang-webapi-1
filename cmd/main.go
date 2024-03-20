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

var sessions = map[string]session{}

type session struct {
	username string
	expiry   time.Time
}

func (s session) isExpired() bool {
	return s.expiry.Before(time.Now())
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
	client *mongo.Client
	err    error
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
		err := errors.New("{\"error\":\"device id must be specified\"}")
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

func Mongo_WriteDevices(devices Root, c *mongo.Client) error {
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

func LoginHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	if r.Method != MethodPost {
		http.Error(w, "{\"error\":\"method not allowed\"}", http.StatusMethodNotAllowed)
		return
	}

	var creds Credentials
	// Get the JSON body and decode into credentials
	err := json.NewDecoder(r.Body).Decode(&creds)
	if err != nil {
		// If the structure of the body is wrong, return an HTTP error
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Get the expected password from our in memory map
	expectedPassword, ok := users[creds.Username]

	// If a password exists for the given user
	// AND, if it is the same as the password we received, the we can move ahead
	// if NOT, then we return an "Unauthorized" status
	if !ok || expectedPassword != creds.Password {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	sessionToken := uuid.NewString()
	expiresAt := time.Now().Add(120 * time.Second)

	sessions[sessionToken] = session{
		username: creds.Username,
		expiry:   expiresAt,
	}

	http.SetCookie(w, &http.Cookie{
		Name:    "session_token",
		Value:   sessionToken,
		Expires: expiresAt,
	})

	client, err = ConnectDB("mongodb://localhost:27017")
	if err != nil {
		log.Fatal("Error connecting to MongoDB: %v\n", err)
	}
}

func CheckSessionHandler(w http.ResponseWriter, r *http.Request) {
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
	w.Write([]byte(fmt.Sprintf("Authorized %s", userSession.username)))
}

func RefreshTokenHandler(w http.ResponseWriter, r *http.Request) {
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
	// (END) The code until this point is the same as the first part of the `Welcome` route

	// If the previous session is valid, create a new session token for the current user
	newSessionToken := uuid.NewString()
	expiresAt := time.Now().Add(120 * time.Second)

	// Set the token in the session map, along with the user whom it represents
	sessions[newSessionToken] = session{
		username: userSession.username,
		expiry:   expiresAt,
	}

	// Delete the older session token
	delete(sessions, sessionToken)

	// Set the new token as the users `session_token` cookie
	http.SetCookie(w, &http.Cookie{
		Name:    "session_token",
		Value:   newSessionToken,
		Expires: time.Now().Add(120 * time.Second),
	})
}

func LogoutHandler(w http.ResponseWriter, r *http.Request, client *mongo.Client) error {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	if r.Method != MethodPost {
		http.Error(w, "{\"error\":\"method not allowed\"}", http.StatusMethodNotAllowed)
		return errors.New("{\"error\":\"method not allowed\"}")
	}

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
		Name:    "session_token",
		Value:   "",
		Expires: time.Now(),
	})

	err = ClientStatusDB(client)
	if err != nil {
		http.Error(w, "{\"error\":\"client not authenticated\"}", http.StatusUnauthorized)
		return err
	}

	client.Disconnect(context.TODO())
	return nil
}

func GetDevicesHandler(w http.ResponseWriter, r *http.Request, client *mongo.Client) error {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	if r.Method != MethodGet {
		http.Error(w, "{\"error\":\"method not allowed\"}", http.StatusMethodNotAllowed)
		return err
	}

	err := ClientStatusDB(client)
	if err != nil {
		http.Error(w, "{\"error\":\"client not authenticated\"}", http.StatusUnauthorized)
		return err
	}

	var devices Root
	devices, err = GetDeviceDB(bson.D{{}}, client)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return err
	}
	alldevices, err := json.Marshal(devices)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return err
	}
	w.Write(alldevices)
	return nil
}

func GetDeviceByIDHandler(w http.ResponseWriter, r *http.Request, client *mongo.Client) error {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	if r.Method != MethodGet {
		http.Error(w, "{\"error\":\"method not allowed\"}", http.StatusMethodNotAllowed)
		return nil
	}

	var devices Root
	err := ClientStatusDB(client)
	if err != nil {
		http.Error(w, "{\"error\":\"client not authenticated\"}", http.StatusUnauthorized)
		return err
	}

	id := r.URL.Query().Get("id")
	if id == "" {
		http.Error(w, "{\"error\":\"device id must be specified\"}", http.StatusBadRequest)
		return nil
	}

	devices, err = GetDeviceDB(primitive.D{{Key: "_id", Value: id}}, client)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return err
	}
	singledevice, err := json.Marshal(devices)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return err
	}
	w.Write(singledevice)
	return nil
}

func DeleteDeviceHandler(w http.ResponseWriter, r *http.Request, client *mongo.Client) error {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	if r.Method != MethodDelete {
		http.Error(w, "{\"error\":\"method not allowed\"}", http.StatusMethodNotAllowed)
		return nil
	}

	err := ClientStatusDB(client)
	if err != nil {
		http.Error(w, "{\"error\":\"client not authenticated\"}", http.StatusUnauthorized)
		return err
	}
	id := r.URL.Query().Get("id")
	if id == "" {
		http.Error(w, "{\"error\":\"device id must be specified\"}", http.StatusBadRequest)
		return nil
	}

	err = DeleteDeviceDB(primitive.D{{Key: "_id", Value: id}}, client)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return nil
	}
	return nil
}

func DeleteDevicesHandler(w http.ResponseWriter, r *http.Request, client *mongo.Client) error {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	if r.Method != MethodPut {
		http.Error(w, "{\"error\":\"method not allowed\"}", http.StatusMethodNotAllowed)
		return err
	}

	err := ClientStatusDB(client)
	if err != nil {
		http.Error(w, "{\"error\":\"client not authenticated\"}", http.StatusUnauthorized)
		return err
	}

	err = DeleteDevicesDB(bson.D{{}}, client)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return err
	}
	return nil
}

func UpsertDevicesHandler(w http.ResponseWriter, r *http.Request, client *mongo.Client) error {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	var devices Root
	if r.Method != MethodPost {
		http.Error(w, "{\"error\":\"method not allowed\"}", http.StatusMethodNotAllowed)
		return nil
	}

	err := ClientStatusDB(client)
	if err != nil {
		http.Error(w, "{\"error\":\"client not authenticated\"}", http.StatusUnauthorized)
		return err
	}

	err = json.NewDecoder(r.Body).Decode(&devices)

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return nil
	}

	if err := Mongo_WriteDevices(devices, client); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return nil
	}
	return nil
}

func main() {
	http.HandleFunc("/login", func(w http.ResponseWriter, r *http.Request) {
		LoginHandler(w, r)
	})
	http.HandleFunc("/logout", func(w http.ResponseWriter, r *http.Request) {
		LogoutHandler(w, r, client)
	})
	http.HandleFunc("/devices", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == MethodGet {
			GetDevicesHandler(w, r, client)
		} else if r.Method == MethodPost {
			UpsertDevicesHandler(w, r, client)
		} else if r.Method == MethodPut {
			DeleteDevicesHandler(w, r, client)
		} else {
			http.Error(w, "{\"error\":\"method not allowed\"}", http.StatusMethodNotAllowed)
		}
	})
	http.HandleFunc("/device", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == MethodGet {
			GetDeviceByIDHandler(w, r, client)
		} else if r.Method == MethodDelete {
			DeleteDeviceHandler(w, r, client)
		} else {
			http.Error(w, "{\"error\":\"method not allowed\"}", http.StatusMethodNotAllowed)
		}

	})

	http.HandleFunc("/session", func(w http.ResponseWriter, r *http.Request) {
		CheckSessionHandler(w, r)
	})
	http.HandleFunc("/refresh", func(w http.ResponseWriter, r *http.Request) {
		RefreshTokenHandler(w, r)
	})

	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatal(err)
	}
}
