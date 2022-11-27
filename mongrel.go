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
	uri        url.URL
	client     *mongo.Client
	ctx        context.Context
	cancel     context.CancelFunc
	collection *mongo.Collection
}

// Connect function to mongo that accepts MongoConnectDetails struct
func (mongoConnectDetails *MongoConnectDetails) Connect() error {
	/*
	   Connect to my mongo
	*/
	// mongodb://mongodb0.example.com:27017
	// mongodb://myDBReader:D1fficultP%40ssw0rd@mongodb0.example.com:27017/?authSource=admin
	var err error
	mongoConnectDetails.ctx, mongoConnectDetails.cancel = context.WithTimeout(context.Background(), 20*time.Second)
	// use defer for the cancel
	defer mongoConnectDetails.cancel()
	hostWithPort := strings.Join(
		[]string{
			mongoConnectDetails.Host,
			strconv.Itoa(mongoConnectDetails.Port),
		}, ":")
	mongodbUser := url.UserPassword(mongoConnectDetails.Username, mongoConnectDetails.Password)
	queryValues := url.Values{}
	queryValues.Set("authSource", mongoConnectDetails.AuthSource)

	mongoConnectDetails.uri = url.URL{
		Scheme:   "mongodb",
		Host:     hostWithPort,
		User:     mongodbUser,
		RawQuery: queryValues.Encode(),
		Path:     mongoConnectDetails.App,
	}

	clientOptions := options.Client().ApplyURI(mongoConnectDetails.uri.String())

	mongoConnectDetails.client, err = mongo.Connect(mongoConnectDetails.ctx, clientOptions)
	if err != nil {
		return err
	}

	// test connection
	err = mongoConnectDetails.client.Ping(mongoConnectDetails.ctx, readpref.Primary())
	if err != nil {
		log.Fatal("Error with pinging db :: try whitelisting current IP address within clust config ", err)
		mongoConnectDetails.Disconnect()
		return err
	}
	fmt.Println("Backend wired to " + mongoConnectDetails.App + " DB...")

	return nil
}

func (mongoConnectDetails *MongoConnectDetails) Disconnect() error {
	err := mongoClient.Disconnect(context.TODO())

	if err != nil {
		return err
	}
	fmt.Println("Connection to database closed.")
	return nil
}

func (mongoConnectDetails *MongoConnectDetails) AssignCollection(dbName string, collectionName string) {
	mongoConnectDetails.collection = mongoConnectDetails.client.Database(dbName).Collection(collectionName)
}

// ListDatabases returns database list
func (mongoConnectDetails *MongoConnectDetails) ListDatabases() ([]string, error) {
	databases, err := mongoConnectDetails.client.ListDatabaseNames(mongoConnectDetails.ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	return databases, nil
}

// GetDocument will retrieve a single experience post by its uuid
func (mongoConnectDetails *MongoConnectDetails) GetDocument(documnetID string, model *interface{}) (*interface{}, error) {
	// create context for db
	mongoConnectDetails.ctx, mongoConnectDetails.cancel = context.WithTimeout(context.Background(), 20*time.Second)
	defer mongoConnectDetails.cancel()

	// mongodb find params
	findOptions := options.FindOne()
	filterInterface := bson.M{"uuid": documnetID}
	err := mongoConnectDetails.collection.FindOne(mongoConnectDetails.ctx, filterInterface, findOptions).Decode(&model)
	if err != nil {
		return nil, err
	}

	return model, nil
}

// CreateDocument
func (mongoConnectDetails *MongoConnectDetails) CreateDocument(document *interface{}) (interface{}, error) {
	// create context for db insert
	mongoConnectDetails.ctx, mongoConnectDetails.cancel = context.WithTimeout(context.Background(), 20*time.Second)
	defer mongoConnectDetails.cancel()

	result, err := mongoConnectDetails.collection.InsertOne(mongoConnectDetails.ctx, document)
	if err != nil {
		return nil, err
	}

	return result.InsertedID, nil
}
