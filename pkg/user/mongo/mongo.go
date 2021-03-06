package mongo

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/rgynn/klottr/pkg/config"
	"github.com/rgynn/klottr/pkg/user"
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

func NewRepository(cfg *config.Config, client *mongo.Client) (user.Repository, error) {

	if cfg == nil {
		return nil, errors.New("no cfg *config.Config provided")
	}

	if client == nil {
		return nil, errors.New("no client *mongo.Client provided")
	}

	return &Repository{
		database:   cfg.DatabaseName,
		collection: "users",
		cfg:        cfg,
		client:     client,
	}, nil
}

func (repo *Repository) Create(ctx context.Context, m *user.Model) error {

	if m == nil {
		return errors.New("no m *user.Model provided")
	}

	m.Votes = user.Votes{
		Threads:  map[string]int8{},
		Comments: map[string]int8{},
	}

	ctx, cancel := context.WithTimeout(ctx, repo.cfg.RequestTimeout)
	defer cancel()

	_, err := repo.client.Database(repo.database).Collection(repo.collection).InsertOne(ctx, m)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return user.ErrAlreadyExists
		}
		return err
	}

	return nil
}

func (repo *Repository) Search(ctx context.Context, username, role *string, from, size int64) ([]*user.Model, error) {

	if role == nil {
		return nil, errors.New("no role provided")
	}

	filter := bson.D{
		primitive.E{Key: "role", Value: *role},
	}

	if username != nil {
		filter = append(filter, primitive.E{Key: "username", Value: *username})
	}

	ctx, cancel := context.WithTimeout(ctx, repo.cfg.RequestTimeout)
	defer cancel()

	cursor, err := repo.client.Database(repo.database).Collection(repo.collection).Find(ctx, filter, options.Find().SetSkip(from).SetLimit(size))
	if err != nil {
		return nil, err
	}

	result := []*user.Model{}
	if err := cursor.All(ctx, &result); err != nil {
		return nil, err
	}

	return result, nil
}

func (repo *Repository) GetByID(ctx context.Context, id *string) (*user.Model, error) {

	if id == nil {
		return nil, errors.New("no id provided")
	}

	ctx, cancel := context.WithTimeout(ctx, repo.cfg.RequestTimeout)
	defer cancel()

	var result *user.Model
	if err := repo.client.Database(repo.database).Collection(repo.collection).FindOne(ctx, bson.D{primitive.E{Key: "_id", Value: *id}}).Decode(&result); err != nil {
		switch err {
		case mongo.ErrNoDocuments:
			return nil, user.ErrNotFound
		default:
			return nil, err
		}
	}

	return result, nil
}

func (repo *Repository) GetByUsername(ctx context.Context, username *string) (*user.Model, error) {

	if username == nil {
		return nil, errors.New("no username provided")
	}

	ctx, cancel := context.WithTimeout(ctx, repo.cfg.RequestTimeout)
	defer cancel()

	var result *user.Model
	if err := repo.client.Database(repo.database).Collection(repo.collection).FindOne(ctx, bson.D{primitive.E{Key: "username", Value: *username}}).Decode(&result); err != nil {
		return nil, err
	}

	return result, nil
}

func (repo *Repository) Deactivate(ctx context.Context, username, role *string) error {

	if username == nil {
		return errors.New("no username provided")
	}

	if role == nil {
		return errors.New("no role provided")
	}

	ctx, cancel := context.WithTimeout(ctx, repo.cfg.RequestTimeout)
	defer cancel()

	res, err := repo.client.Database(repo.database).Collection(repo.collection).UpdateOne(ctx,
		bson.D{
			primitive.E{Key: "role", Value: *role},
			primitive.E{Key: "username", Value: *username},
		},
		bson.D{primitive.E{
			Key: "$set",
			Value: bson.D{primitive.E{
				Key:   "deactivated",
				Value: time.Now().UTC(),
			}},
		}},
	)
	if err != nil {
		return err
	}

	if res.ModifiedCount != 1 {
		return user.ErrNotFound
	}

	return nil
}

func (repo *Repository) Delete(ctx context.Context, username, role *string) error {

	if username == nil {
		return errors.New("no username provided")
	}

	if role == nil {
		return errors.New("no role provided")
	}

	ctx, cancel := context.WithTimeout(ctx, repo.cfg.RequestTimeout)
	defer cancel()

	res, err := repo.client.Database(repo.database).Collection(repo.collection).DeleteOne(ctx, bson.D{
		primitive.E{Key: "role", Value: *role},
		primitive.E{Key: "username", Value: *username},
	})
	if err != nil {
		return err
	}

	if res.DeletedCount != 1 {
		return user.ErrNotFound
	}

	return nil
}

func (repo *Repository) IncCounter(ctx context.Context, username, field *string, value int8) error {

	if username == nil {
		return errors.New("no username provided")
	}

	if field == nil {
		return errors.New("no field provided")
	}

	ctx, cancel := context.WithTimeout(ctx, repo.cfg.RequestTimeout)
	defer cancel()

	res, err := repo.client.Database(repo.database).Collection(repo.collection).UpdateOne(ctx,
		bson.D{primitive.E{
			Key: "username", Value: *username,
		}},
		bson.D{primitive.E{
			Key: "$inc",
			Value: bson.D{
				primitive.E{Key: *field, Value: value},
			},
		}},
	)
	if err != nil {
		return err
	}

	if res.MatchedCount != 1 {
		return user.ErrNotFound
	}

	return nil
}

func (repo *Repository) UpsertVote(ctx context.Context, username *string, vote *user.Vote) error {

	if username == nil {
		return errors.New("no username provided")
	}

	if vote == nil {
		return errors.New("no vote provided")
	}

	ctx, cancel := context.WithTimeout(ctx, repo.cfg.RequestTimeout)
	defer cancel()

	res, err := repo.client.Database(repo.database).Collection(repo.collection).UpdateOne(ctx,
		bson.D{primitive.E{
			Key: "username", Value: *username,
		}},
		bson.D{primitive.E{
			Key: "$set",
			Value: bson.D{
				primitive.E{Key: fmt.Sprintf("votes.%s.%s", *vote.SlugType, *vote.SlugID), Value: *vote.Value},
			},
		}},
	)
	if err != nil {
		return err
	}

	if res.MatchedCount != 1 {
		return user.ErrNotFound
	}

	return nil
}
