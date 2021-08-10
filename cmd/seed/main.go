package main

import (
	"context"
	"log"

	"github.com/rgynn/klottr/pkg/config"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func main() {

	cfg, err := config.NewFromEnv()
	if err != nil {
		log.Fatal(err)
	}

	client, err := openDB(cfg)
	if err != nil {
		log.Fatal(err)
	}

	// todo: make sure thread collections are capped with expiration on documents
	// todo: make sure indexes are present on collections

	if err := closeDB(cfg, client); err != nil {
		log.Fatal(err)
	}
}

func openDB(cfg *config.Config) (*mongo.Client, error) {

	ctx, cancel := context.WithTimeout(context.Background(), cfg.RequestTimeout)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(cfg.DatabaseURL))
	if err != nil {
		log.Fatal(err)
	}

	if err := client.Ping(ctx, nil); err != nil {
		log.Fatal(err)
	}

	log.Printf("INFO: Connected to database with uri: %s\n", cfg.DatabaseURL)

	return client, nil
}

func closeDB(cfg *config.Config, client *mongo.Client) error {

	ctx, cancel := context.WithTimeout(context.Background(), cfg.RequestTimeout)
	defer cancel()

	if err := client.Disconnect(ctx); err != nil {
		return err
	}

	log.Printf("INFO: Disconnected from database\n")

	return nil
}
