package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"

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

const (
	MethodGet    = "GET"
	MethodPost   = "POST"
	MethodPut    = "PUT"
	MethodDelete = "DELETE"
)

var (
	client *mongo.Client
	err    error
)

func Mongo_ConnectDB(mongoURI string) (*mongo.Client, error) {
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

func MongoClientStatus(client *mongo.Client) error {
	if client == nil {
		return errors.New("error: no client connection to database")
	}
	return nil
}

func Mongo_GetDevice(filter bson.D, c *mongo.Client) (Root, error) {
	var devices Root
	err := MongoClientStatus(client)
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

func Mongo_DeleteDevice(filter bson.D, c *mongo.Client) error {
	err := MongoClientStatus(client)
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

func Mongo_DeleteAllDevices(filter bson.D, c *mongo.Client) error {
	err := MongoClientStatus(client)
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
	err := MongoClientStatus(client)
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

	client, err = Mongo_ConnectDB("mongodb://localhost:27017")
	if err != nil {
		log.Fatal("Error connecting to MongoDB: %v\n", err)
	}
}

func LogoutHandler(w http.ResponseWriter, r *http.Request, client *mongo.Client) error {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	if r.Method != MethodPost {
		http.Error(w, "{\"error\":\"method not allowed\"}", http.StatusMethodNotAllowed)
		return errors.New("{\"error\":\"method not allowed\"}")
	}

	err := MongoClientStatus(client)
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

	err := MongoClientStatus(client)
	if err != nil {
		http.Error(w, "{\"error\":\"client not authenticated\"}", http.StatusUnauthorized)
		return err
	}

	var devices Root
	devices, err = Mongo_GetDevice(bson.D{{}}, client)
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
	err := MongoClientStatus(client)
	if err != nil {
		http.Error(w, "{\"error\":\"client not authenticated\"}", http.StatusUnauthorized)
		return err
	}

	id := r.URL.Query().Get("id")
	if id == "" {
		http.Error(w, "{\"error\":\"device id must be specified\"}", http.StatusBadRequest)
		return nil
	}

	devices, err = Mongo_GetDevice(primitive.D{{Key: "_id", Value: id}}, client)
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

func DeleteSingleDeviceHandler(w http.ResponseWriter, r *http.Request, client *mongo.Client) error {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	if r.Method != MethodDelete {
		http.Error(w, "{\"error\":\"method not allowed\"}", http.StatusMethodNotAllowed)
		return nil
	}

	err := MongoClientStatus(client)
	if err != nil {
		http.Error(w, "{\"error\":\"client not authenticated\"}", http.StatusUnauthorized)
		return err
	}
	id := r.URL.Query().Get("id")
	if id == "" {
		http.Error(w, "{\"error\":\"device id must be specified\"}", http.StatusBadRequest)
		return nil
	}

	err = Mongo_DeleteDevice(primitive.D{{Key: "_id", Value: id}}, client)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return nil
	}
	return nil
}

func DeleteAllDevicesHandler(w http.ResponseWriter, r *http.Request, client *mongo.Client) error {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	if r.Method != MethodPut {
		http.Error(w, "{\"error\":\"method not allowed\"}", http.StatusMethodNotAllowed)
		return err
	}

	err := MongoClientStatus(client)
	if err != nil {
		http.Error(w, "{\"error\":\"client not authenticated\"}", http.StatusUnauthorized)
		return err
	}

	err = Mongo_DeleteAllDevices(bson.D{{}}, client)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return err
	}
	return nil
}

func PostDevicesHandler(w http.ResponseWriter, r *http.Request, client *mongo.Client) error {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	var devices Root
	if r.Method != MethodPost {
		http.Error(w, "{\"error\":\"method not allowed\"}", http.StatusMethodNotAllowed)
		return nil
	}

	err := MongoClientStatus(client)
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
	http.HandleFunc("/createdevice", func(w http.ResponseWriter, r *http.Request) {
		PostDevicesHandler(w, r, client)
	})
	http.HandleFunc("/getdevices", func(w http.ResponseWriter, r *http.Request) {
		GetDevicesHandler(w, r, client)
	})
	http.HandleFunc("/getdevicebyid", func(w http.ResponseWriter, r *http.Request) {
		GetDeviceByIDHandler(w, r, client)
	})
	http.HandleFunc("/deletedevice", func(w http.ResponseWriter, r *http.Request) {
		DeleteSingleDeviceHandler(w, r, client)
	})
	http.HandleFunc("/cleardevices", func(w http.ResponseWriter, r *http.Request) {
		DeleteAllDevicesHandler(w, r, client)
	})

	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatal(err)
	}
}
