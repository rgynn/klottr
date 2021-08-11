package main

import (
	"context"
	"fmt"

	"github.com/rgynn/klottr/pkg/config"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var logger = logrus.New()
var threadCategories = []string{"misc"}

func main() {

	cfg, err := config.NewFromEnv()
	if err != nil {
		logger.Fatal(err)
	}

	client, err := openDB(cfg)
	if err != nil {
		logger.Fatal(err)
	}

	if err := createCappedThreadsCollections(cfg, client); err != nil {
		logger.Fatal(err)
	}

	if err := createCommentsCollection(cfg, client); err != nil {
		logger.Fatal(err)
	}

	if err := createUsersCollection(cfg, client); err != nil {
		logger.Fatal(err)
	}

	if err := closeDB(cfg, client); err != nil {
		logger.Fatal(err)
	}
}

func createCappedThreadsCollections(cfg *config.Config, client *mongo.Client) error {

	ctx, cancel := context.WithTimeout(context.Background(), cfg.RequestTimeout)
	defer cancel()

	for _, category := range threadCategories {

		name := fmt.Sprintf("threads_%s", category)

		logger.Infof("Dropping collection: %s in database: %s", name, cfg.DatabaseName)
		if err := client.Database(cfg.DatabaseName).Collection(name).Drop(ctx); err != nil {
			logger.Warn(err)
		}

		logger.Infof("Creating collection: %s in database: %s", name, cfg.DatabaseName)
		if err := client.Database(cfg.DatabaseName).CreateCollection(ctx, name,
			options.CreateCollection().
				SetCapped(true).
				SetSizeInBytes(1000000000),
		); err != nil {
			return err
		}

		logger.Infof("Creating indexes for collection: %s in database: %s", name, cfg.DatabaseName)
		indexes, err := client.Database(cfg.DatabaseName).Collection(name).Indexes().CreateMany(ctx,
			[]mongo.IndexModel{
				{
					Keys: bson.D{
						primitive.E{Key: "_id", Value: 1},
					},
				},
				{
					Keys: bson.D{
						primitive.E{Key: "slug_id", Value: 1},
					},
					Options: options.Index().SetUnique(true),
				},
				{
					Keys: bson.D{
						primitive.E{Key: "category", Value: 1},
						primitive.E{Key: "slug_id", Value: 1},
						primitive.E{Key: "slug_title", Value: 1},
					},
				},
				{
					Keys: bson.D{
						primitive.E{Key: "user_id", Value: 1},
					},
				},
				{
					Keys: bson.D{
						primitive.E{Key: "created", Value: 1},
					},
					Options: options.Index().SetExpireAfterSeconds(cfg.PostTTLSeconds),
				},
			},
		)
		if err != nil {
			return err
		}

		for _, idx := range indexes {
			logger.Infof("Created index: %s for collection: %s", idx, name)
		}
	}

	return nil
}

func createCommentsCollection(cfg *config.Config, client *mongo.Client) error {

	ctx, cancel := context.WithTimeout(context.Background(), cfg.RequestTimeout)
	defer cancel()

	name := "comments"

	logger.Infof("Dropping collection: %s in database: %s", name, cfg.DatabaseName)
	if err := client.Database(cfg.DatabaseName).Collection(name).Drop(ctx); err != nil {
		logger.Warn(err)
	}

	logger.Infof("Creating collection: %s in database: %s", name, cfg.DatabaseName)
	if err := client.Database(cfg.DatabaseName).CreateCollection(ctx, name); err != nil {
		return err
	}

	logger.Infof("Creating indexes for collection: %s in database: %s", name, cfg.DatabaseName)
	indexes, err := client.Database(cfg.DatabaseName).Collection(name).Indexes().CreateMany(ctx,
		[]mongo.IndexModel{
			{
				Keys: bson.D{
					primitive.E{Key: "_id", Value: 1},
				},
			},
			{
				Keys: bson.D{
					primitive.E{Key: "thread_id", Value: 1},
					primitive.E{Key: "user_id", Value: 1},
				},
			},
		},
	)
	if err != nil {
		return err
	}

	for _, idx := range indexes {
		logger.Infof("Created index: %s for collection: %s", idx, name)
	}

	return nil
}

func createUsersCollection(cfg *config.Config, client *mongo.Client) error {

	ctx, cancel := context.WithTimeout(context.Background(), cfg.RequestTimeout)
	defer cancel()

	name := "comments"

	logger.Infof("Dropping collection: %s in database: %s", name, cfg.DatabaseName)
	if err := client.Database(cfg.DatabaseName).Collection(name).Drop(ctx); err != nil {
		logger.Warn(err)
	}

	logger.Infof("Creating collection: %s in database: %s", name, cfg.DatabaseName)
	if err := client.Database(cfg.DatabaseName).CreateCollection(ctx, name); err != nil {
		return err
	}

	logger.Infof("Creating indexes for collection: %s in database: %s", name, cfg.DatabaseName)
	indexes, err := client.Database(cfg.DatabaseName).Collection(name).Indexes().CreateMany(ctx,
		[]mongo.IndexModel{
			{
				Keys: bson.D{
					primitive.E{Key: "_id", Value: 1},
				},
			},
			{
				Keys: bson.D{
					primitive.E{Key: "username", Value: 1},
				},
				Options: options.Index().SetUnique(true),
			},
			{
				Keys: bson.D{
					primitive.E{Key: "username", Value: 1},
					primitive.E{Key: "role", Value: 1},
				},
			},
		},
	)
	if err != nil {
		return err
	}

	for _, idx := range indexes {
		logger.Infof("Created index: %s for collection: %s", idx, name)
	}

	return nil
}

func openDB(cfg *config.Config) (*mongo.Client, error) {

	ctx, cancel := context.WithTimeout(context.Background(), cfg.RequestTimeout)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(cfg.DatabaseURL))
	if err != nil {
		return nil, err
	}

	if err := client.Ping(ctx, nil); err != nil {
		return nil, err
	}

	logger.Infof("INFO: Connected to database with uri: %s\n", cfg.DatabaseURL)

	return client, nil
}

func closeDB(cfg *config.Config, client *mongo.Client) error {

	ctx, cancel := context.WithTimeout(context.Background(), cfg.RequestTimeout)
	defer cancel()

	if err := client.Disconnect(ctx); err != nil {
		return err
	}

	logger.Infof("INFO: Disconnected from database\n")

	return nil
}
