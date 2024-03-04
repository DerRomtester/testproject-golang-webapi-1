package main

import (
    "fmt"
    "context"
    "log"
    "encoding/json"
    "net/http"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Root struct {
	Devices []struct {
		ID                              string `json:"id"`
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
	} `json:"devices"`
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

var MongoDb = *Mongo_ConnectDB()

func Mongo_ConnectDB() *mongo.Client {

    	mongoURI := MongoURI
    	client, err := mongo.Connect(context.TODO(), options.Client().ApplyURI(mongoURI))
	if err != nil {
		log.Fatal(err)
		panic(err)
	}
	defer func() {
		if err := client.Disconnect(context.TODO()); err != nil {
			log.Fatal(err)
			panic(err)
		}
	}()
	fmt.Println("Connected to MongoDb")
	return client
}

func Mongo_WriteDevices(devices Root, c* mongo.Client) {
	collection := c.Database("devices-db").Collection("Devices")
	docs := []interface{}{
		devices.Devices,
	}
	fmt.Print(docs...)
	result, err := collection.InsertMany(context.TODO(), docs)
	if err != nil {
		log.Fatal(err)
		panic(err)
	}

	fmt.Printf("Documents inserted: %v\n", len(result.InsertedIDs))

	for _, id := range result.InsertedIDs {
		fmt.Printf("Inserted document with _id: %v\n", id)
	}
}

func PostDevicesHandler(w http.ResponseWriter, r *http.Request) {
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
    Mongo_WriteDevices(devices, &MongoDb)
    fmt.Fprintf(w, "Devices: %+v", devices.Devices[0].ID)
}

func main() {
    	http.HandleFunc("/postDevices", PostDevicesHandler)
    	http.ListenAndServe(":8080", nil)
}
