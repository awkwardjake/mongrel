package mongrel

import (
	"context"
	"fmt"
	"log"
	"net/url"
	"strconv"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

var mongoClient *mongo.Client

// MongoConnectDetails model
type MongoConnectDetails struct {
	Username   string `json:"username"`
	Password   string `json:"password"`
	Host       string `json:"host"`
	Port       int    `json:"port"`
	AuthSource string `json:"authSource"`
	App        string `json:"app"`
}

// Connect function to mongo that accepts MongoConnectDetails struct
func Connect(mongoConnectDetails *MongoConnectDetails) (*mongo.Client, error) {
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
		Path:     mongoConnectDetails.App,
	}

	clientOptions := options.Client().ApplyURI(mongodbURI.String())

	mongoClient, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		return nil, err
	}

	// test connection
	err = mongoClient.Ping(ctx, readpref.Primary())
	if err != nil {
		log.Fatal("Error with pinging db :: try whitelisting current IP address within clust config ", err)
		Disconnect()
	}
	fmt.Println("Backend wired to " + mongoConnectDetails.App + " DB...")
	/*
	   List databases
	*/
	databases, err := mongoClient.ListDatabaseNames(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	fmt.Println("Current Mongo Databases: ", databases)

	return mongoClient, nil
}

func Disconnect() error {
	err := mongoClient.Disconnect(context.TODO())

	if err != nil {
		return err
	}
	fmt.Println("Connection to MongoDB closed.")
	return nil
}

func AssignCollection(mongoClient *mongo.Client, dbName string, collectionName string) *mongo.Collection {
	return mongoClient.Database(dbName).Collection(collectionName)
}

// GetDocument will retrieve a single experience post by its uuid
func GetDocument(collection *mongo.Collection, documnetID string, model *interface{}) (*interface{}, error) {
	// create context for db
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	// mongodb find params
	findOptions := options.FindOne()
	filterInterface := bson.M{"uuid": documnetID}
	err := collection.FindOne(ctx, filterInterface, findOptions).Decode(&model)
	if err != nil {
		return nil, err
	}

	return model, nil
}

// CreateDocument
func CreateDocument(collection *mongo.Collection, document *interface{}) (interface{}, error) {
	// create context for db insert
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	result, err := collection.InsertOne(ctx, document)
	if err != nil {
		fmt.Println("error inserting document into database :: ", err)
		return "", err
	}

	return result.InsertedID, nil
}
