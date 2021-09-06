package db

import (
	"context"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"healthcare-system-sawtooth/client/lib"
	"log"
	"sync"
	"sync/atomic"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	// Name of the table in MongoDB
	MongoDataCollection    = "Datas"
	MongoRequestCollection = "Requests"
)

// MongoDB connection client
var mongoClient *mongo.Client

// Data read/write race condition prevention variables
var mongoInit uint32
var mongoMu sync.Mutex

// GetMongoDbName retrieves the database name to use
func GetMongoDbName() string {
	return lib.MongoDbName
}

// GetMongoClient gets the singleton MongoDB client
func GetMongoClient(ctx context.Context) (*mongo.Client, error) {
	// We're purposely not using sync.Once here. This is because
	// connections to mongdb can temporarily fail, and we want
	// to keep retrying
	if atomic.LoadUint32(&mongoInit) == 1 {
		return mongoClient, nil
	}

	mongoMu.Lock()
	defer mongoMu.Unlock()

	if mongoInit == 0 {
		client, err := newMongoClient(ctx,
			lib.MongoDbUrl)
		if err != nil {
			return nil, err
		}

		err = CreateIndexes(ctx, client)
		if err != nil {
			client.Disconnect(ctx)
			return nil, err
		}

		mongoClient = client

		atomic.StoreUint32(&mongoInit, 1)
	}

	return mongoClient, nil
}

func noop() {}

// GetMongoContext creates a new mongodb context
func GetMongoContext() (context.Context, context.CancelFunc) {
	var ctx context.Context
	var cancel context.CancelFunc
	mongoTimeout := 0

	if mongoTimeout == 0 {
		ctx = context.Background()
		cancel = noop
	} else {
		ctx, cancel = context.WithTimeout(context.Background(),
			time.Duration(mongoTimeout)*time.Second)
	}

	return ctx, cancel
}

// Get table name
func GetMongoCollection(ctx context.Context, colName string) (*mongo.Collection, error) {
	client, err := GetMongoClient(ctx)
	if err != nil {
		return nil, err
	}
	return client.Database(GetMongoDbName()).Collection(colName), nil
}

// Creates indexes for better search
func CreateIndexes(ctx context.Context, client *mongo.Client) error {
	col := client.Database(GetMongoDbName()).Collection(MongoDataCollection)

	models := []mongo.IndexModel{
		{
			Keys:    bson.M{"hash": 1},
			Options: options.Index().SetUnique(true),
		},
	}

	_, err := col.Indexes().CreateMany(ctx, models)
	if err != nil {
		return err
	}

	return nil
}

// Creates MongoDB connection client
func newMongoClient(ctx context.Context, url string) (*mongo.Client, error) {
	// Set client options
	clientOptions := options.Client().ApplyURI(url)

	// Connect to MongoDB
	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		return nil, err
	}

	// Check the connection
	err = client.Ping(ctx, nil)
	if err != nil {
		return nil, err
	}

	log.Println("Connected to MongoDB!")
	return client, nil
}

func GetMongoObject(ctx context.Context, collection *mongo.Collection, id *primitive.ObjectID) *mongo.SingleResult {
	return collection.FindOne(ctx, bson.M{"_id": id})
}
