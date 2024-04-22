package database

import (
	"context"
	"errors"
	"log"

	"github.com/DerRomtester/testproject-golang-webapi-1/model"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func ConnectDB(db model.DatabaseConnection) (*mongo.Client, error) {
	ctx, cancel := context.WithTimeout(context.Background(), db.Timeout)
	defer cancel()
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(db.GetConnStr()))

	if err != nil {
		log.Fatal("failed to create mongo client")
		return nil, err
	}

	err = client.Ping(ctx, nil)
	if err != nil {
		log.Fatal("Error pinging database")
		return nil, err
	}

	log.Println("Connected to db host: ", db.Host, " Port: ", db.Port)
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

func DeleteDeviceDB(filter bson.D, client *mongo.Client, deleteMany bool) error {
	err := ClientStatusDB(client)
	if err != nil {
		return err
	}

	collection := client.Database("devices-db").Collection("Devices")
	if !deleteMany && filter == nil {	
			err := errors.New("device id must be specified")
			return err
	} 

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

func CreateUserDB(user model.UserCredentials, client *mongo.Client) error {
	err := ClientStatusDB(client)
	if err != nil {
		return err
	}

	collection := client.Database("users-db").Collection("users")
	filter := bson.D{primitive.E{Key: "_id", Value: user.Username}}
	var existingUser model.UserCredentials
	err = collection.FindOne(context.Background(), filter).Decode(&existingUser)
	if err == nil {
		update := bson.D{primitive.E{Key: "$set", Value: bson.M{"username": user.Username}}}
		_, err := collection.UpdateOne(context.Background(), filter, update)
		if err != nil {
			return err
		}
	} else {
		_, err := collection.InsertOne(context.Background(), user)
		if err != nil {
			return err
		}
	}
	return nil
}

func GetUserDB(username string, client *mongo.Client) (model.UserCredentials, error) {
	var user model.UserCredentials
	filter := bson.D{primitive.E{Key: "username", Value: username}}
	err := ClientStatusDB(client)
	if err != nil {
		return user, err
	}

	collection := client.Database("users-db").Collection("users")
	err = collection.FindOne(context.Background(), filter).Decode(&user)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return user, err
		}
		return user, err
	}
	return user, nil
}

func CreateSessionDB(s model.UserSession, client *mongo.Client) error {
	err := ClientStatusDB(client)
	if err != nil {
		return err
	}

	collection := client.Database("users-db").Collection("session")
	filter := bson.D{primitive.E{Key: "token", Value: s.Token}}
	var existingToken model.UserSession
	err = collection.FindOne(context.Background(), filter).Decode(&existingToken)
	if err == nil {
		update := bson.D{primitive.E{Key: "$set", Value: bson.M{"token": s.Token}}}
		_, err := collection.UpdateOne(context.Background(), filter, update)
		if err != nil {
			return err
		}
	} else {
		_, err := collection.InsertOne(context.Background(), s)
		if err != nil {
			return err
		}
	}
	return nil
}

func GetTokenDB(token string, client *mongo.Client) (model.UserSession, error) {
	var session model.UserSession
	filter := bson.D{primitive.E{Key: "token", Value: token}}
	err := ClientStatusDB(client)
	if err != nil {
		return session, err
	}

	collection := client.Database("users-db").Collection("session")
	err = collection.FindOne(context.Background(), filter).Decode(&session)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return session, err
		}
		return session, err
	}
	return session, nil
}

func DeleteTokenDB(token string, client *mongo.Client) error {
	filter := bson.D{primitive.E{Key: "token", Value: token}}

	err := ClientStatusDB(client)
	if err != nil {
		return err
	}

	collection := client.Database("users-db").Collection("session")
	_, err = collection.DeleteOne(context.TODO(), filter)
	if err != nil {
		return err
	}
	return nil
}
