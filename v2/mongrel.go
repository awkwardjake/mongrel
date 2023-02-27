package mongrel

import (
	"context"
	"fmt"
	"log"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

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

const defaultTimeout = 6 * time.Second

// Connect function to mongo that accepts MongoConnectDetails struct
func (mongoConnectDetails *MongoConnectDetails) Connect(requestContext context.Context) error {
	/*
	   Connect to my mongo
	*/
	// mongodb://mongodb0.example.com:27017
	// mongodb://myDBReader:D1fficultP%40ssw0rd@mongodb0.example.com:27017/?authSource=admin
	var err error
	mongoConnectDetails.ctx, mongoConnectDetails.cancel = context.WithTimeout(requestContext, defaultTimeout)
	// use defer for the cancel
	defer mongoConnectDetails.cancel()

	hostWithPort := strings.Join(
		[]string{
			mongoConnectDetails.Host,
			strconv.Itoa(mongoConnectDetails.Port),
		}, ":")

	credentials := options.Credential{
		AuthMechanism: "SCRAM-SHA-256",
		AuthSource:    mongoConnectDetails.App,
		Username:      mongoConnectDetails.Username,
		Password:      mongoConnectDetails.Password,
	}

	mongoConnectDetails.uri = url.URL{
		Scheme: "mongodb",
		Host:   hostWithPort,
		Path:   mongoConnectDetails.App,
	}

	clientOptions := options.Client().ApplyURI(mongoConnectDetails.uri.String()).SetAuth(credentials)

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
	err := mongoConnectDetails.client.Disconnect(mongoConnectDetails.ctx)

	if err != nil {
		return err
	}
	fmt.Println("Connection to database closed.")
	return nil
}

func (mongoConnectDetails *MongoConnectDetails) SelectCollection(dbName string, collectionName string) {
	mongoConnectDetails.collection = mongoConnectDetails.client.Database(dbName).Collection(collectionName)
}

// ListDatabases returns database list
func (mongoConnectDetails *MongoConnectDetails) ListDatabases(requestContext context.Context) ([]string, error) {
	mongoConnectDetails.ctx, mongoConnectDetails.cancel = context.WithTimeout(requestContext, defaultTimeout)
	defer mongoConnectDetails.cancel()
	databases, err := mongoConnectDetails.client.ListDatabaseNames(mongoConnectDetails.ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	return databases, nil
}

// CreateDocument
func (mongoConnectDetails *MongoConnectDetails) CreateDocument(requestContext context.Context, document interface{}) (interface{}, error) {
	mongoConnectDetails.ctx, mongoConnectDetails.cancel = context.WithTimeout(requestContext, defaultTimeout)
	defer mongoConnectDetails.cancel()

	_, err := mongoConnectDetails.collection.InsertOne(mongoConnectDetails.ctx, document)
	if err != nil {
		return nil, err
	}
	if err != nil {
		return nil, err
	}
	return document, nil
}

func (mongoConnectDetails *MongoConnectDetails) UpdateDocumentByID(requestContext context.Context, field string, id string, update interface{}) (int, error) {
	mongoConnectDetails.ctx, mongoConnectDetails.cancel = context.WithTimeout(requestContext, defaultTimeout)
	defer mongoConnectDetails.cancel()
	filter := bson.M{field: id}
	updateData := bson.D{{"$set", update}}
	result, err := mongoConnectDetails.collection.UpdateOne(requestContext, filter, updateData)
	if err != nil {
		return 0, err
	}
	log.Println(result)
	return int(result.ModifiedCount), nil
}

func (mongoConnectDetails *MongoConnectDetails) InsertManyDocuments(requestContext context.Context, documents []interface{}) (int, error) {
	mongoConnectDetails.ctx, mongoConnectDetails.cancel = context.WithTimeout(requestContext, defaultTimeout)
	defer mongoConnectDetails.cancel()

	result, err := mongoConnectDetails.collection.InsertMany(mongoConnectDetails.ctx, documents)
	if err != nil {
		return 0, err
	}

	return len(result.InsertedIDs), nil
}

// find documents

// GetDocument will retrieve a single document by provided "id" field
// returned interface{} value can be Marshalled and Unmarshalled as needed
func (mongoConnectDetails *MongoConnectDetails) GetDocumentByID(requestContext context.Context, field string, id string) (interface{}, error) {
	// create context for db
	mongoConnectDetails.ctx, mongoConnectDetails.cancel = context.WithTimeout(requestContext, defaultTimeout)
	defer mongoConnectDetails.cancel()
	var result interface{}
	filter := bson.M{field: id}
	err := mongoConnectDetails.collection.FindOne(mongoConnectDetails.ctx, filter).Decode(&result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

// CreateIndex expects a single field and optionall a desired sort order ("asc" or "desc")
// for nested fields, pass a string that traverses fields using "."
func (mongoConnectDetails *MongoConnectDetails) CreateIndex(requestContext context.Context, field string, opts ...string) error {
	descendingSort, err := regexp.Compile(`(?i)desc`)
	if err != nil {
		return err
	}
	ascendingSort, err := regexp.Compile(`(?i)asc`)
	if err != nil {
		return err
	}
	textIndex, err := regexp.Compile(`(?i)text`)
	if err != nil {
		return err
	}

	var idxModel mongo.IndexModel

	if len(opts) > 0 {
		// asc = 1
		// desc = -1
		if descendingSort.Match([]byte(opts[0])) {
			sortOption := -1
			idxModel.Keys = bson.D{{field, sortOption}}
		}
		if ascendingSort.Match([]byte(opts[0])) {
			sortOption := 1
			idxModel.Keys = bson.D{{field, sortOption}}
		}
		if textIndex.Match([]byte(opts[0])) {
			idxModel.Keys = bson.D{{field, "text"}}
		}
	}
	mongoConnectDetails.ctx, mongoConnectDetails.cancel = context.WithTimeout(requestContext, defaultTimeout)
	defer mongoConnectDetails.cancel()
	result, err := mongoConnectDetails.collection.Indexes().CreateOne(mongoConnectDetails.ctx, idxModel)
	if err != nil {
		return err
	}

	log.Println(result)
	return nil
}

// SearchDocument expects request context and relies on a field name and single term string to find documents
func (mongoConnectDetails *MongoConnectDetails) SearchDocumentsByField(requestContext context.Context, field string, term string) ([]interface{}, error) {
	mongoConnectDetails.ctx, mongoConnectDetails.cancel = context.WithTimeout(requestContext, defaultTimeout)
	defer mongoConnectDetails.cancel()
	var results []interface{}
	filter := bson.D{{"$text", bson.D{{"$search", term}}}}
	cursor, err := mongoConnectDetails.collection.Find(mongoConnectDetails.ctx, filter)

	if err != nil {
		return nil, err
	}

	if err = cursor.All(mongoConnectDetails.ctx, &results); err != nil {
		return nil, err
	}

	return results, nil
}

// DeleteDocument expects request context in addition to specifying the field and identifier for the target document to delete
func (mongoConnectDetails *MongoConnectDetails) DeleteDocument(requestContext context.Context, field string, id string) (int, error) {
	mongoConnectDetails.ctx, mongoConnectDetails.cancel = context.WithTimeout(requestContext, defaultTimeout)
	defer mongoConnectDetails.cancel()

	filter := bson.M{field: id}
	result, err := mongoConnectDetails.collection.DeleteOne(mongoConnectDetails.ctx, filter)
	if err != nil {
		return 0, err
	}
	return int(result.DeletedCount), nil
}
