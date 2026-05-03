package database

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

func DBSet() (*mongo.Client, error) {
	// MongoDB connection setup logic goes here
	clientOptions := options.Client().ApplyURI("mongodb://root:password@localhost:27017/?retryWrites=true&w=majority") // Replace with your MongoDB URI
	clientOptions.SetServerSelectionTimeout(10 * time.Second)
	clientOptions.SetConnectTimeout(10 * time.Second) // Set server selection timeout

	minPoolSize := uint64(20)
	maxPoolSize := uint64(100)
	clientOptions.SetMinPoolSize(minPoolSize)
	clientOptions.SetMaxPoolSize(maxPoolSize)

	const maxRetries = 5
	baseDelay := 500 * time.Millisecond
	maxDelay := 30 * time.Second

	var lastErr error

	for attempt := 0; attempt <= maxRetries; attempt++ {
		var retryErr error
		client, err := mongo.Connect(clientOptions)
		if err == nil {
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			pingErr := client.Ping(ctx, nil)
			cancel()
			if pingErr == nil {
				fmt.Println("Connected to MongoDB!")
				return client, nil
			}
			_ = client.Disconnect(context.Background()) // Clean up the client if ping fails
			retryErr = pingErr
		} else {
			retryErr = err
		}

		lastErr = retryErr // Capture the last connection/ping error
		if attempt == maxRetries {
			break
		}

		delay := baseDelay * time.Duration(1<<attempt) // Exponential backoff
		if delay > maxDelay {
			delay = maxDelay
		}
		log.Printf("MongoDB unavailable (attempt %d/%d): %v. Retrying in %s",
			attempt+1, maxRetries+1, retryErr, delay)
		time.Sleep(delay)
	}
	log.Fatalf("MongoDB unavailable after %d attempts: %v", maxRetries+1, lastErr)
	return nil, lastErr
}

func UserData(client *mongo.Client, collectionName, databaseName string) *mongo.Collection {
	var collection *mongo.Collection = client.Database(databaseName).Collection(collectionName)
	if collection == nil {
		log.Fatalf("Failed to get collection: %s", collectionName)
	}

	return collection
}

func ProductData(client *mongo.Client, collectionName, databaseName string) *mongo.Collection {
	var productCollection *mongo.Collection = client.Database(databaseName).Collection(collectionName)
	if productCollection == nil {
		log.Fatalf("Failed to get collection: %s", collectionName)
	}

	return productCollection
}

func EnsureCollections(ctx context.Context, client *mongo.Client, databaseName string, collectionNames ...string) error {
	if len(collectionNames) == 0 {
		return nil
	}

	db := client.Database(databaseName)

	existingNames, err := db.ListCollectionNames(ctx, map[string]string{})
	if err != nil {
		return err
	}

	existing := make(map[string]struct{}, len(existingNames))
	for _, name := range existingNames {
		existing[name] = struct{}{}
	}

	for _, name := range collectionNames {
		if _, ok := existing[name]; ok {
			continue
		}

		err = db.CreateCollection(ctx, name)
		if err != nil {
			var commandErr mongo.CommandError
			if errors.As(err, &commandErr) && commandErr.Code == 48 {
				continue
			}
			return err
		}
	}

	return nil
}
