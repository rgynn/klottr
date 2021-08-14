package mongo

import (
	"context"
	"errors"

	"github.com/rgynn/klottr/pkg/comment"
	"github.com/rgynn/klottr/pkg/config"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Repository struct {
	database   string
	collection string
	cfg        *config.Config
	client     *mongo.Client
}

func NewRepository(cfg *config.Config, client *mongo.Client) (comment.Repository, error) {

	if cfg == nil {
		return nil, errors.New("no cfg *config.Config provided")
	}

	if client == nil {
		return nil, errors.New("no client *mongo.Client provided")
	}

	return &Repository{
		database:   cfg.DatabaseName,
		collection: "comments",
		cfg:        cfg,
		client:     client,
	}, nil
}

func (repo *Repository) Create(ctx context.Context, m *comment.Model) error {

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

func (repo *Repository) Get(ctx context.Context, slugID *string) (*comment.Model, error) {

	if slugID == nil {
		return nil, errors.New("no slugID provided")
	}

	ctx, cancel := context.WithTimeout(ctx, repo.cfg.RequestTimeout)
	defer cancel()

	var result *comment.Model
	if err := repo.client.Database(repo.database).Collection(repo.collection).FindOne(ctx, bson.D{
		primitive.E{Key: "slug_id", Value: *slugID},
	}).Decode(&result); err != nil {
		return nil, err
	}

	return result, nil
}

func (repo *Repository) ListByThreadID(ctx context.Context, threadID *primitive.ObjectID, from, size int64) ([]*comment.Model, error) {

	if threadID == nil {
		return nil, errors.New("no theadID provided")
	}

	ctx, cancel := context.WithTimeout(ctx, repo.cfg.RequestTimeout)
	defer cancel()

	cursor, err := repo.client.Database(repo.database).Collection(repo.collection).Find(ctx, bson.D{
		primitive.E{Key: "thread_id", Value: *threadID},
	}, options.Find().SetSkip(from).SetLimit(size))
	if err != nil {
		return nil, err
	}

	result := []*comment.Model{}
	if err := cursor.All(ctx, &result); err != nil {
		return nil, err
	}

	return result, nil
}

func (repo *Repository) ListByUsername(ctx context.Context, username *string, from, size int64) ([]*comment.Model, error) {

	if username == nil {
		return nil, errors.New("no username provided")
	}

	ctx, cancel := context.WithTimeout(ctx, repo.cfg.RequestTimeout)
	defer cancel()

	cursor, err := repo.client.Database(repo.database).Collection(repo.collection).Find(ctx, bson.D{
		primitive.E{Key: "username", Value: *username},
	}, options.Find().SetSkip(from).SetLimit(size))
	if err != nil {
		return nil, err
	}

	result := []*comment.Model{}
	if err := cursor.All(ctx, &result); err != nil {
		return nil, err
	}

	return result, nil
}

func (repo *Repository) Delete(ctx context.Context, slugID *string) error {

	if slugID == nil {
		return errors.New("no slugID provided")
	}

	ctx, cancel := context.WithTimeout(ctx, repo.cfg.RequestTimeout)
	defer cancel()

	res, err := repo.client.Database(repo.database).Collection(repo.collection).DeleteOne(ctx, bson.D{
		primitive.E{Key: "slug_id", Value: *slugID},
	})
	if err != nil {
		return err
	}

	if res.DeletedCount != 1 {
		return comment.ErrNotFound
	}

	return nil
}

func (repo *Repository) IncVotes(ctx context.Context, slugID *string) error {

	if slugID == nil {
		return errors.New("no slugID provided")
	}

	ctx, cancel := context.WithTimeout(ctx, repo.cfg.RequestTimeout)
	defer cancel()

	res, err := repo.client.Database(repo.database).Collection(repo.collection).UpdateOne(ctx, bson.D{
		primitive.E{Key: "slug_id", Value: *slugID},
	}, bson.D{
		primitive.E{
			Key: "$inc",
			Value: bson.D{
				primitive.E{Key: "votes", Value: 1},
			},
		}})
	if err != nil {
		return err
	}

	if res.ModifiedCount != 1 {
		return comment.ErrNotFound
	}

	return nil
}

func (repo *Repository) DecVotes(ctx context.Context, slugID *string) error {

	if slugID == nil {
		return errors.New("no slugID provided")
	}

	ctx, cancel := context.WithTimeout(ctx, repo.cfg.RequestTimeout)
	defer cancel()

	res, err := repo.client.Database(repo.database).Collection(repo.collection).UpdateOne(ctx, bson.D{
		primitive.E{Key: "slug_id", Value: *slugID},
	}, bson.D{
		primitive.E{
			Key: "$inc",
			Value: bson.D{
				primitive.E{Key: "votes", Value: -1},
			},
		}})
	if err != nil {
		return err
	}

	if res.ModifiedCount != 1 {
		return comment.ErrNotFound
	}

	return nil
}
