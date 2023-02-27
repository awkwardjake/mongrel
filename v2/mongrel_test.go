package mongrel

import (
	"context"
	"fmt"
	"log"
	"testing"

	uuid "github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
)

var testMongrel = MongoConnectDetails{
	Username:   "appUser",
	Password:   "appUserPass",
	Host:       "localhost",
	Port:       27017,
	AuthSource: "test",
	App:        "test",
}

type TestDoc struct {
	UUID    string `bson:"id" json:"id"`
	Name    string `bson:"name" json:"name"`
	Message string `bson:"message" json:"message"`
}

func generateID() string {
	id := uuid.New()
	return id.String()
}

var documentID string

func TestMongrel_Connect(test *testing.T) {

	err := testMongrel.Connect(context.Background())

	if err != nil {
		test.Errorf("unable to connect to test mongrel DB: %s", err)
	}

	defer testMongrel.Disconnect()
}

func TestMongrel_CreateDocument(test *testing.T) {
	err := testMongrel.Connect(context.Background())

	if err != nil {
		log.Fatalf("unable to connect to test mongrel DB to create document: %s", err)
	}

	defer testMongrel.Disconnect()
	doc := TestDoc{
		UUID:    generateID(),
		Name:    "My Test",
		Message: "Hello there!",
	}
	// assign/create collection in db
	// create document
	testMongrel.SelectCollection("test", "testCollection")
	_, err = testMongrel.CreateDocument(context.Background(), doc)
	if err != nil {
		test.Errorf("unable to create document: %s", err)
	}

	documentID = doc.UUID
}

func TestMongrel_UpdateDocumentByID(test *testing.T) {
	err := testMongrel.Connect(context.Background())
	if err != nil {
		log.Fatalf("unable to connect to test mongrel DB to update document by ID: %s", err)
	}
	defer testMongrel.Disconnect()

	updatedDoc := TestDoc{
		UUID:    documentID,
		Name:    "My updated doc",
		Message: "Hello, I've been updated",
	}
	log.Printf("updated %s", documentID)
	testMongrel.SelectCollection("test", "testCollection")
	numberUpdated, err := testMongrel.UpdateDocumentByID(context.Background(), "id", documentID, updatedDoc)
	if err != nil || !(numberUpdated > 0) {
		test.Errorf("didn't update document as expected: %s", err)
	}
}

func TestMongrel_InsertManyDocuments(test *testing.T) {
	err := testMongrel.Connect(context.Background())
	if err != nil {
		log.Fatalf("unable to connect to DB to insert many documents: %s", err)
	}

	defer testMongrel.Disconnect()
	newDocs := make([]interface{}, 2)
	for idx := range newDocs {
		newDocs[idx] = TestDoc{
			UUID:    generateID(),
			Name:    fmt.Sprintf("doc%v", idx+1),
			Message: fmt.Sprintf("doc%v message", idx+1),
		}
	}
	testMongrel.SelectCollection("test", "testCollection")
	numberInserted, err := testMongrel.InsertManyDocuments(context.Background(), newDocs)
	if err != nil || !(numberInserted > 0) {
		test.Errorf("could not insert many documents: %s", err)
	}
	log.Printf("inserted %v documents", numberInserted)
}

func TestMongrel_ListDatabases(test *testing.T) {
	err := testMongrel.Connect(context.Background())

	if err != nil {
		log.Fatalf("unable to connect to DB to list DBs: %s", err)
	}

	defer testMongrel.Disconnect()

	dbList, err := testMongrel.ListDatabases(context.Background())
	if err != nil && !(len(dbList) > 0) {
		test.Errorf("error listing dbs or no dbs available: %s", err)
	}
}

func TestMongrel_GetDocumentByID(test *testing.T) {
	err := testMongrel.Connect(context.Background())

	if err != nil {
		log.Fatalf("unable to connect to DB to get document: %s", err)
	}

	defer testMongrel.Disconnect()

	testMongrel.SelectCollection("test", "testCollection")

	result, err := testMongrel.GetDocumentByID(context.Background(), "id", documentID)
	if err != nil {
		test.Errorf("error getting document by ID: %s", err)
	}
	newDoc, err := bson.Marshal(result)
	if err != nil {
		test.Errorf("error marshaling document by ID: %s", err)
	}
	var testDoc TestDoc
	err = bson.Unmarshal(newDoc, &testDoc)

	if err != nil || testDoc.UUID != documentID {
		test.Errorf("error unmarshaling or getting correct document by ID: %s", err)
	}
}

func TestMongrel_CreateIndex(test *testing.T) {
	err := testMongrel.Connect(context.Background())

	if err != nil {
		log.Fatalf("unable to connect to DB to create index of field: %s", err)
	}
	defer testMongrel.Disconnect()

	testMongrel.SelectCollection("test", "testCollection")
	err = testMongrel.CreateIndex(context.Background(), "name", "text")
	if err != nil {
		test.Errorf("couldn't create index: %s", err)
	}
}

func TestMongrel_SearchDocumentsByField(test *testing.T) {
	err := testMongrel.Connect(context.Background())

	if err != nil {
		log.Fatalf("unable to connect to DB to search documents: %s", err)
	}

	defer testMongrel.Disconnect()

	testMongrel.SelectCollection("test", "testCollection")
	results, err := testMongrel.SearchDocumentsByField(context.Background(), "name", "doc1")
	if err != nil || !(len(results) > 0) {
		test.Errorf("error searching for documents: %s", err)
	}
	var testDocs []TestDoc
	for _, result := range results {
		newDoc, err := bson.Marshal(result)
		if err != nil {
			test.Errorf("error marshalling results: %s", err)
		}
		var testDoc TestDoc
		err = bson.Unmarshal(newDoc, &testDoc)
		if err != nil {
			test.Errorf("error unmarshalling results: %s", err)
		}
		testDocs = append(testDocs, testDoc)
	}

	log.Println(testDocs)
}

func TestMongrel_DeleteDocument(test *testing.T) {
	err := testMongrel.Connect(context.Background())

	if err != nil {
		log.Fatalf("unable to connect to DB to delete document: %s", err)
	}
	defer testMongrel.Disconnect()

	testMongrel.SelectCollection("test", "testCollection")
	numberDeleted, err := testMongrel.DeleteDocument(context.Background(), "id", documentID)
	log.Printf("number deleted %v", numberDeleted)
	if err != nil || !(numberDeleted > 0) {
		test.Errorf("unable to delete anything: %s", err)
	}
}
