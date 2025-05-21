package controller

import (
	"context"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const dbName = "authdb"
const collectionName = "profile"

var client *mongo.Client

var collection *mongo.Collection

func init() {
	const mongoURI = "mongodb+srv://Sarthak:Sarthak%402910@book-store.shs1wxp.mongodb.net/?retryWrites=true&w=majority&appName=Book-Store"
	if mongoURI == "" {
		log.Fatal("MONGO_URI not set in .env")
	}
	clientOptions := options.Client().ApplyURI(mongoURI)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var err error
	client, err = mongo.Connect(ctx, clientOptions)
	if err != nil {
		log.Fatalf("Failed to connect to MongoDB: %v", err)
	}

	err = client.Ping(ctx, nil)
	if err != nil {
		log.Fatalf("Failed to ping MongoDB: %v", err)
	}

	log.Println("MongoDB connection successful")
	collection = client.Database(dbName).Collection(collectionName)
	log.Println("Collection instance created")
}

func GetMongoClient() *mongo.Client {
	if client == nil {
		log.Fatal("MongoDB client is nil; initialization failed")
	}
	return client
}
func GetCollection() *mongo.Collection {
	if collection == nil {
		log.Fatal("MongoDB collection is nil; initialization failed")
	}
	return collection
}