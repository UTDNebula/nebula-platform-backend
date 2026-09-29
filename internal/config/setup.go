package config

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

// Closure to execute the function that connects to MongoDB only once
var ConnectMongo = sync.OnceValues(func() (*mongo.Client, error) {
	mongoUri, err := GetEnvMongoURI()
	if err != nil {
		return nil, err
	}

	opts := options.Client().ApplyURI(mongoUri)
	client, err := mongo.Connect(opts)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := client.Ping(ctx, nil); err != nil {
		client.Disconnect(context.Background())
		return nil, fmt.Errorf("Unable to ping database: %v", err)
	}

	log.Printf("Connected to MongoDB")
	return client, nil
})

func GetCollection(name string) (*mongo.Collection, error) {
	client, err := ConnectMongo()
	if err != nil {
		return nil, err
	}
	collection := client.Database("combinedDB").Collection(name)
	return collection, nil
}
