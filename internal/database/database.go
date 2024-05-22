package database

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/DerRomtester/testproject-golang-webapi-1/internal/helper"
	"github.com/DerRomtester/testproject-golang-webapi-1/internal/session"
	"github.com/DerRomtester/testproject-golang-webapi-1/model"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	ErrCreateMongoClient = "failed to create mongo client"
	ErrPingDB            = "error pinging database"
	ErrNoClient          = "no client connection"
	ErrDeviceID          = "no device id specified"
)

type DBClient struct {
	Client *mongo.Client
}

type DatabaseConnection struct {
	User     string
	Password string
	Timeout  time.Duration
	Host     string
	Port     string
}

func (db DatabaseConnection) GetConnStr() string {
	if db.User == "" || db.Password == "" {
		return fmt.Sprintf("mongodb://%s:%s", db.Host, db.Port)
	} else {
		return fmt.Sprintf("mongodb://%s:%s@%s:%s", db.User, db.Password, db.Host, db.Port)
	}
}

func (db DatabaseConnection) ConnectDB() (DBClient, error) {
	ctx, cancel := context.WithTimeout(context.Background(), db.Timeout)
	defer cancel()
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(db.GetConnStr()))

	if err != nil {
		log.Fatal(ErrCreateMongoClient)
		return DBClient{}, err
	}

	err = client.Ping(ctx, nil)
	if err != nil {
		log.Fatal(ErrPingDB)
		return DBClient{}, err
	}

	log.Println("Connected to db host: ", db.Host, " Port: ", db.Port)
	return DBClient{Client: client}, nil
}

func (mg DBClient) ClientStatusDB() error {
	if mg.Client == nil {
		return errors.New(ErrNoClient)
	}
	return nil
}

func (mg DBClient) GetDeviceDB(filter bson.D) (model.Devices, error) {
	var devices model.Devices
	err := mg.ClientStatusDB()
	if err != nil {
		return devices, err
	}

	collection := mg.Client.Database("devices-db").Collection("Devices")
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

func (mg DBClient) DeleteDeviceDB(filter bson.D, deleteMany bool) error {
	err := mg.ClientStatusDB()
	if err != nil {
		return err
	}

	collection := mg.Client.Database("devices-db").Collection("Devices")
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

func (mg DBClient) WriteDevicesDB(devices model.Devices) error {
	err := mg.ClientStatusDB()
	if err != nil {
		return err
	}

	collection := mg.Client.Database("devices-db").Collection("Devices")
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

func (mg DBClient) CheckUserExists(username string) error {
	var (
		err          error
		existingUser model.UserCredentials
	)

	err = mg.ClientStatusDB()
	if err != nil {
		return err
	}

	if username == "" {
		return errors.New("username is empty")
	}

	collection := mg.Client.Database("users-db").Collection("users")
	filter := bson.D{primitive.E{Key: "username", Value: username}}
	err = collection.FindOne(context.Background(), filter).Decode(&existingUser)

	if err != nil {
		return err
	} else {
		return errors.New("user found in db")
	}
}

func (mg DBClient) CreateUserDB(user model.UserCredentials) error {
	err := mg.ClientStatusDB()
	if err != nil {
		return err
	}

	err = mg.CheckUserExists(user.Username)
	if err.Error() == "user found in db" {
		return err
	}

	if (err.Error() != "user found in db") && (err.Error() != "mongo: no documents in result") && (err != nil) {
		return err
	}

	collection := mg.Client.Database("users-db").Collection("users")
	hashedPassword, err := helper.HashPassword(user.Password)

	if err != nil {
		return err
	}

	_, err = collection.InsertOne(context.Background(), model.UserCredentials{Username: user.Username, Password: hashedPassword})
	if err != nil {
		return err
	}

	return nil
}

func (mg DBClient) GetUserDB(username string) (model.UserCredentials, error) {
	var (
		err  error
		user model.UserCredentials
	)

	filter := bson.D{primitive.E{Key: "username", Value: username}}
	err = mg.ClientStatusDB()
	if err != nil {
		return user, err
	}

	collection := mg.Client.Database("users-db").Collection("users")
	err = collection.FindOne(context.Background(), filter).Decode(&user)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return user, err
		}
		return user, err
	}
	return user, nil
}

func (mg DBClient) CreateSessionDB(s session.UserSession) error {
	err := mg.ClientStatusDB()
	if err != nil {
		return err
	}

	collection := mg.Client.Database("users-db").Collection("session")
	filter := bson.D{primitive.E{Key: "token", Value: s.Token}}
	err = collection.FindOne(context.Background(), filter).Decode(s.GetToken())
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

func (mg DBClient) GetTokenDB(token string) (session.UserSession, error) {
	var session session.UserSession
	filter := bson.D{primitive.E{Key: "token", Value: token}}
	err := mg.ClientStatusDB()
	if err != nil {
		return session, err
	}

	collection := mg.Client.Database("users-db").Collection("session")
	err = collection.FindOne(context.Background(), filter).Decode(&session)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return session, err
		}
		return session, err
	}
	return session, nil
}

func (mg DBClient) DeleteTokenDB(token string) error {
	filter := bson.D{primitive.E{Key: "token", Value: token}}

	err := mg.ClientStatusDB()
	if err != nil {
		return err
	}

	collection := mg.Client.Database("users-db").Collection("session")
	_, err = collection.DeleteOne(context.TODO(), filter)
	if err != nil {
		return err
	}
	return nil
}
