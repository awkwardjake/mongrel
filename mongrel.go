package mongrel

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var mongoClient *mongo.Client

// MongoConnectDetails model
type MongoConnectDetails struct {
	Username   string `json:"username"`
	Password   string `json:"password"`
	Host       string `json:"host"`
	Port       int    `json:"port"`
	AuthSource string `json:"authSource"`
}

// Connect function to mongo that accepts MongoConnectDetails struct
func Connect(mongoConnectDetails *MongoConnectDetails) error {
	/*
	   Connect to my mongo
	*/
	// mongodb://mongodb0.example.com:27017
	// mongodb://myDBReader:D1fficultP%40ssw0rd@mongodb0.example.com:27017/?authSource=admin

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	// use defer for the cancel
	defer cancel()
	hostWithPort := strings.Join(
		[]string{
			mongoConnectDetails.Host,
			strconv.Itoa(mongoConnectDetails.Port),
		}, ":")
	mongodbUser := url.UserPassword(mongoConnectDetails.Username, mongoConnectDetails.Password)
	queryValues := url.Values{}
	queryValues.Set("authSource", mongoConnectDetails.AuthSource)

	mongodbURI := url.URL{
		Scheme:   "mongodb",
		Host:     hostWithPort,
		User:     mongodbUser,
		RawQuery: queryValues.Encode(),
		Path:     "/",
	}

	clientOptions := options.Client().ApplyURI(mongodbURI.String())

	mongoClient, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		return err
	}
	/*
	   List databases
	*/
	databases, err := mongoClient.ListDatabaseNames(ctx, bson.M{})
	if err != nil {
		return err
	}
	fmt.Println("Current Mongo Databases: ", databases)
	return nil
}

func Disconnect() error {
	err := mongoClient.Disconnect(context.TODO())

	if err != nil {
		return err
	}
	fmt.Println("Connection to MongoDB closed.")
	return nil
}
