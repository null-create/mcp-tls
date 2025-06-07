package db

import (
	"context"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// RPCMessage represents a JSON-RPC message structure
type RPCMessage struct {
	ID        string                 `bson:"_id,omitempty" json:"id"`
	Method    string                 `bson:"method" json:"method"`
	Params    map[string]interface{} `bson:"params" json:"params"`
	Result    interface{}            `bson:"result,omitempty" json:"result,omitempty"`
	Error     interface{}            `bson:"error,omitempty" json:"error,omitempty"`
	Timestamp time.Time              `bson:"timestamp" json:"timestamp"`
}

// MongoStore is a wrapper for MongoDB client and collection
type MongoStore struct {
	client     *mongo.Client
	collection *mongo.Collection
	ctx        context.Context
	cancel     context.CancelFunc
}

// NewMongoStore creates and returns a new MongoStore
func NewMongoStore(uri, dbName, collectionName string) (*MongoStore, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)

	clientOpts := options.Client().ApplyURI(uri)
	client, err := mongo.Connect(ctx, clientOpts)
	if err != nil {
		cancel()
		return nil, err
	}

	// Ping to ensure the connection is live
	if err := client.Ping(ctx, nil); err != nil {
		cancel()
		return nil, err
	}

	collection := client.Database(dbName).Collection(collectionName)

	store := &MongoStore{
		client:     client,
		collection: collection,
		ctx:        ctx,
		cancel:     cancel,
	}

	log.Println("✅ Connected to MongoDB")
	return store, nil
}

// InsertMessage inserts a new JSON-RPC message into the database
func (ms *MongoStore) InsertMessage(msg RPCMessage) error {
	msg.Timestamp = time.Now()
	_, err := ms.collection.InsertOne(ms.ctx, msg)
	return err
}

// FindMessagesByMethod retrieves messages filtered by method name
func (ms *MongoStore) FindMessagesByMethod(method string) ([]RPCMessage, error) {
	filter := bson.M{"method": method}
	cursor, err := ms.collection.Find(ms.ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ms.ctx)

	var messages []RPCMessage
	if err := cursor.All(ms.ctx, &messages); err != nil {
		return nil, err
	}
	return messages, nil
}

// Close disconnects the MongoDB client
func (ms *MongoStore) Close() error {
	ms.cancel()
	return ms.client.Disconnect(context.Background())
}

// // example usage
// func main() {
// 	store, err := NewMongoStore("mongodb://localhost:27017", "rpc_db", "messages")
// 	if err != nil {
// 		log.Fatal(err)
// 	}
// 	defer store.Close()
// }
