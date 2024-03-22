package database

import (
	"context"
	"errors"
	"fmt"
	"github.com/DerRomtester/testproject-golang-webapi-1/model"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
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

func GetDeviceDB(filter bson.D, client *mongo.Client) (model.Root, error) {
	var devices model.Root
	err := ClientStatusDB(client)
	if err != nil {
		return devices, err
	}

	collection := client.Database("devices-db").Collection("Devices")
	cursor, err := collection.Find(context.Background(), filter)
	if err != nil {
		return devices, err
	}
	defer cursor.Close(context.Background())

	for cursor.Next(context.Background()) {
		var device model.Device
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

func DeleteDeviceDB(filter bson.D, client *mongo.Client) error {
	err := ClientStatusDB(client)
	if err != nil {
		return err
	}

	collection := client.Database("devices-db").Collection("Devices")
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

func DeleteDevicesDB(filter bson.D, client *mongo.Client) error {
	err := ClientStatusDB(client)
	if err != nil {
		return err
	}

	collection := client.Database("devices-db").Collection("Devices")
	_, err = collection.DeleteMany(context.TODO(), filter)
	if err != nil {
		return err
	}
	return nil
}

func WriteDevicesDB(devices model.Root, client *mongo.Client) error {
	err := ClientStatusDB(client)
	if err != nil {
		return err
	}

	collection := client.Database("devices-db").Collection("Devices")
	for _, device := range devices.Devices {
		filter := bson.D{primitive.E{Key: "_id", Value: device.ID}}
		var existingDevice model.Device
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
