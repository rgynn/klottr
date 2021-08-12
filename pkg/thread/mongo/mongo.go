package mongo

import (
	"context"
	"errors"
	"fmt"

	"github.com/rgynn/klottr/pkg/config"
	"github.com/rgynn/klottr/pkg/thread"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Repository for threads in mongo cluster
type Repository struct {
	database   string
	collection string
	cfg        *config.Config
	client     *mongo.Client
}

func NewRepository(cfg *config.Config, category string) (thread.Repository, error) {

	if cfg == nil {
		return nil, errors.New("no cfg *config.Config provided")
	}

	if category == "" {
		return nil, errors.New("must supply a category for thread repisotory")
	}

	ctx, cancel := context.WithTimeout(context.Background(), cfg.RequestTimeout)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(cfg.DatabaseURL))
	if err != nil {
		return nil, err
	}

	if err := client.Ping(ctx, nil); err != nil {
		return nil, err
	}

	return &Repository{
		database:   cfg.DatabaseName,
		collection: fmt.Sprintf("threads_%s", category),
		cfg:        cfg,
		client:     client,
	}, nil
}

func (repo *Repository) List(ctx context.Context, from, size int64) ([]*thread.Model, error) {

	ctx, cancel := context.WithTimeout(ctx, repo.cfg.RequestTimeout)
	defer cancel()

	cursor, err := repo.client.Database(repo.database).Collection(repo.collection).Find(ctx, bson.D{}, options.Find().SetSkip(from).SetLimit(size))
	if err != nil {
		return nil, err
	}

	result := []*thread.Model{}
	if err := cursor.All(ctx, &result); err != nil {
		return nil, err
	}

	return result, nil
}

func (repo *Repository) Create(ctx context.Context, m *thread.Model) error {

	if m == nil {
		return errors.New("no m *thread.Model provided")
	}

	ctx, cancel := context.WithTimeout(ctx, repo.cfg.RequestTimeout)
	defer cancel()

	_, err := repo.client.Database(repo.database).Collection(repo.collection).InsertOne(ctx, m)
	if err != nil {
		return err
	}

	return nil
}

func (repo *Repository) Get(ctx context.Context, slugID, slugTitle *string) (*thread.Model, error) {

	if slugID == nil {
		return nil, errors.New("no slugID provided")
	}

	filter := bson.D{
		primitive.E{Key: "slug_id", Value: *slugID},
	}

	if slugTitle != nil {
		filter = append(filter, primitive.E{Key: "slug_title", Value: *slugTitle})
	}

	ctx, cancel := context.WithTimeout(ctx, repo.cfg.RequestTimeout)
	defer cancel()

	var result *thread.Model
	if err := repo.client.Database(repo.database).Collection(repo.collection).FindOne(ctx, filter).Decode(&result); err != nil {
		return nil, err
	}

	return result, nil
}

func (repo *Repository) Delete(ctx context.Context, slugID, slugTitle *string) error {

	if slugID == nil {
		return errors.New("no slugID provided")
	}

	filter := bson.D{
		primitive.E{Key: "slug_id", Value: *slugID},
	}

	if slugTitle != nil {
		filter = append(filter, primitive.E{Key: "slug_title", Value: *slugTitle})
	}

	ctx, cancel := context.WithTimeout(ctx, repo.cfg.RequestTimeout)
	defer cancel()

	res, err := repo.client.Database(repo.database).Collection(repo.collection).DeleteOne(ctx, filter)
	if err != nil {
		return err
	}

	if res.DeletedCount != 1 {
		return thread.ErrNotFound
	}

	return nil
}

func (repo *Repository) IncVote(ctx context.Context, slugID, slugTitle *string) error {

	if slugID == nil {
		return errors.New("no slugID provided")
	}

	filter := bson.D{
		primitive.E{Key: "slug_id", Value: *slugID},
	}

	if slugTitle != nil {
		filter = append(filter, primitive.E{Key: "slug_title", Value: *slugTitle})
	}

	ctx, cancel := context.WithTimeout(ctx, repo.cfg.RequestTimeout)
	defer cancel()

	res, err := repo.client.Database(repo.database).Collection(repo.collection).UpdateOne(ctx, filter, bson.D{
		primitive.E{
			Key: "$inc",
			Value: bson.D{
				primitive.E{Key: "counters.votes", Value: 1},
			},
		}})
	if err != nil {
		return err
	}

	if res.ModifiedCount != 1 {
		return thread.ErrNotFound
	}

	return nil
}

func (repo *Repository) DecVote(ctx context.Context, slugID, slugTitle *string) error {

	if slugID == nil {
		return errors.New("no slugID provided")
	}

	filter := bson.D{
		primitive.E{Key: "slug_id", Value: *slugID},
	}

	if slugTitle != nil {
		filter = append(filter, primitive.E{Key: "slug_title", Value: *slugTitle})
	}

	ctx, cancel := context.WithTimeout(ctx, repo.cfg.RequestTimeout)
	defer cancel()

	res, err := repo.client.Database(repo.database).Collection(repo.collection).UpdateOne(ctx, filter, bson.D{
		primitive.E{
			Key: "$dec",
			Value: bson.D{
				primitive.E{Key: "counters.votes", Value: 1},
			},
		}})
	if err != nil {
		return err
	}

	if res.ModifiedCount != 1 {
		return thread.ErrNotFound
	}

	return nil
}

func (repo *Repository) IncComments(ctx context.Context, slugID, slugTitle *string) error {

	if slugID == nil {
		return errors.New("no slugID provided")
	}

	filter := bson.D{
		primitive.E{Key: "slug_id", Value: *slugID},
	}

	if slugTitle != nil {
		filter = append(filter, primitive.E{Key: "slug_title", Value: *slugTitle})
	}

	ctx, cancel := context.WithTimeout(ctx, repo.cfg.RequestTimeout)
	defer cancel()

	res, err := repo.client.Database(repo.database).Collection(repo.collection).UpdateOne(ctx, filter, bson.D{
		primitive.E{
			Key: "$inc",
			Value: bson.D{
				primitive.E{Key: "counters.comments", Value: 1},
			},
		}})
	if err != nil {
		return err
	}

	if res.ModifiedCount != 1 {
		return thread.ErrNotFound
	}

	return nil
}

func (repo *Repository) DecComments(ctx context.Context, slugID, slugTitle *string) error {

	if slugID == nil {
		return errors.New("no slugID provided")
	}

	filter := bson.D{
		primitive.E{Key: "slug_id", Value: *slugID},
	}

	if slugTitle != nil {
		filter = append(filter, primitive.E{Key: "slug_title", Value: *slugTitle})
	}

	ctx, cancel := context.WithTimeout(ctx, repo.cfg.RequestTimeout)
	defer cancel()

	res, err := repo.client.Database(repo.database).Collection(repo.collection).UpdateOne(ctx, filter, bson.D{
		primitive.E{
			Key: "$dec",
			Value: bson.D{
				primitive.E{Key: "counters.comments", Value: 1},
			},
		}})
	if err != nil {
		return err
	}

	if res.ModifiedCount != 1 {
		return thread.ErrNotFound
	}

	return nil
}

func (repo *Repository) Close() error {
	ctx, cancel := context.WithTimeout(context.Background(), repo.cfg.RequestTimeout)
	defer cancel()
	return repo.client.Disconnect(ctx)
}
