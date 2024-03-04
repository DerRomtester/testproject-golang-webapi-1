package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

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

type Root struct {
	Devices []Device `json:"devices"`
}

const (
	MethodGet     = "GET"
	MethodHead    = "HEAD"
	MethodPost    = "POST"
	MethodPut     = "PUT"
	MethodPatch   = "PATCH" // RFC 5789
	MethodDelete  = "DELETE"
	MethodConnect = "CONNECT"
	MethodOptions = "OPTIONS"
	MethodTrace   = "TRACE"
	MongoURI      = "mongodb://localhost:27017"
)

func Mongo_ConnectDB() (*mongo.Client, error) {
    	mongoURI := MongoURI
    	client, err := mongo.Connect(context.TODO(), options.Client().ApplyURI(mongoURI))
	if err != nil {
		return nil, err
	}
	
	err = client.Ping(context.Background(), nil)
	if err != nil {
		return nil, err
	}

	fmt.Println("Connected to MongoDb")
	return client, nil
}

func Mongo_WriteDevices(devices Root, c* mongo.Client) error {
	collection := c.Database("devices-db").Collection("Devices")

	for _, device := range devices.Devices {
		filter := bson.D{{"_id", device.ID}}
		var existingDevice Device
		err := collection.FindOne(context.Background(), filter).Decode(&existingDevice)
		if err == nil {
			update := bson.D{{"$set", bson.M{"name": device.Name}}}
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

func PostDevicesHandler(w http.ResponseWriter, r *http.Request, client *mongo.Client) {
    w.Header().Set("Content-Type", "application/json")
    w.Header().Set("Access-Control-Allow-Origin", "*")

    var devices Root
    
    if r.Method != MethodPost {
	    http.Error(w, "" , http.StatusMethodNotAllowed)
	    return
    }

    err := json.NewDecoder(r.Body).Decode(&devices)

    if err != nil {
	http.Error(w, err.Error(), http.StatusBadRequest)
	return
    }

    if err := Mongo_WriteDevices(devices, client); err != nil {
	http.Error(w, err.Error(), http.StatusInternalServerError)
	return
    }
}

func main() {
	client, err := Mongo_ConnectDB()
	if err != nil {
		log.Fatal("Error connecting to MongoDB: %v\n", err)
	}
	defer client.Disconnect(context.Background())

	http.HandleFunc("/postDevices", func(w http.ResponseWriter, r *http.Request) {
		PostDevicesHandler(w, r, client)
	})
    	http.ListenAndServe(":8080", nil)
}
