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

func NewRepository(cfg *config.Config) (comment.Repository, error) {

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

func (repo *Repository) Get(ctx context.Context, id *string) (*comment.Model, error) {

	if id == nil {
		return nil, errors.New("no id provided")
	}

	ctx, cancel := context.WithTimeout(ctx, repo.cfg.RequestTimeout)
	defer cancel()

	var result *comment.Model
	if err := repo.client.Database(repo.database).Collection(repo.collection).FindOne(ctx, bson.D{
		primitive.E{Key: "_id", Value: *id},
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

func (repo *Repository) ListByUserID(ctx context.Context, userID *primitive.ObjectID, from, size int64) ([]*comment.Model, error) {

	if userID == nil {
		return nil, errors.New("no userID provided")
	}

	ctx, cancel := context.WithTimeout(ctx, repo.cfg.RequestTimeout)
	defer cancel()

	cursor, err := repo.client.Database(repo.database).Collection(repo.collection).Find(ctx, bson.D{
		primitive.E{Key: "user_id", Value: *userID},
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

func (repo *Repository) Delete(ctx context.Context, id *string) error {

	if id == nil {
		return errors.New("no id provided")
	}

	ctx, cancel := context.WithTimeout(ctx, repo.cfg.RequestTimeout)
	defer cancel()

	res, err := repo.client.Database(repo.database).Collection(repo.collection).DeleteOne(ctx, bson.D{
		primitive.E{Key: "_id", Value: *id},
	})
	if err != nil {
		return err
	}

	if res.DeletedCount != 1 {
		return comment.ErrNotFound
	}

	return nil
}

func (repo *Repository) IncVotes(ctx context.Context, id *string) error {

	if id == nil {
		return errors.New("no id provided")
	}

	ctx, cancel := context.WithTimeout(ctx, repo.cfg.RequestTimeout)
	defer cancel()

	res, err := repo.client.Database(repo.database).Collection(repo.collection).UpdateOne(ctx, bson.D{
		primitive.E{Key: "_id", Value: *id},
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

func (repo *Repository) DecVotes(ctx context.Context, id *string) error {

	if id == nil {
		return errors.New("no id provided")
	}

	ctx, cancel := context.WithTimeout(ctx, repo.cfg.RequestTimeout)
	defer cancel()

	res, err := repo.client.Database(repo.database).Collection(repo.collection).UpdateOne(ctx, bson.D{
		primitive.E{Key: "_id", Value: *id},
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

func (repo *Repository) Close() error {
	ctx, cancel := context.WithTimeout(context.Background(), repo.cfg.RequestTimeout)
	defer cancel()
	return repo.client.Disconnect(ctx)
}
