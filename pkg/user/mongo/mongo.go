package mongo

import (
	"context"
	"errors"
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

func NewRepository(cfg *config.Config) (user.Repository, error) {

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
		collection: "users",
		cfg:        cfg,
		client:     client,
	}, nil
}

func (repo *Repository) Create(ctx context.Context, m *user.Model) error {

	if m == nil {
		return errors.New("no m *user.Model provided")
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
		bson.D{
			primitive.E{
				Key: "$set",
				Value: bson.D{primitive.E{
					Key:   "deactivated",
					Value: time.Now().UTC()},
				},
			},
		},
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

func (repo *Repository) IncThreadsCounter(ctx context.Context, username *string) error {

	if username == nil {
		return errors.New("no username provided")
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
				primitive.E{Key: "counters.num.threads", Value: 1},
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

func (repo *Repository) DecThreadsCounter(ctx context.Context, username *string) error {

	if username == nil {
		return errors.New("no username provided")
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
				primitive.E{Key: "counters.num.threads", Value: -1},
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

func (repo *Repository) IncCommentsCounter(ctx context.Context, username *string) error {

	if username == nil {
		return errors.New("no username provided")
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
				primitive.E{Key: "counters.num.comments", Value: 1},
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

func (repo *Repository) DecCommentsCounter(ctx context.Context, username *string) error {

	if username == nil {
		return errors.New("no username provided")
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
				primitive.E{Key: "counters.num.comments", Value: -1},
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

func (repo *Repository) IncThreadsVotes(ctx context.Context, username *string) error {

	if username == nil {
		return errors.New("no username provided")
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
				primitive.E{Key: "counters.votes.threads", Value: 1},
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

func (repo *Repository) DecThreadsVotes(ctx context.Context, username *string) error {

	if username == nil {
		return errors.New("no username provided")
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
				primitive.E{Key: "counters.votes.threads", Value: -1},
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

func (repo *Repository) IncCommentsVotes(ctx context.Context, username *string) error {

	if username == nil {
		return errors.New("no username provided")
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
				primitive.E{Key: "counters.votes.comments", Value: 1},
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

func (repo *Repository) DecCommentsVotes(ctx context.Context, username *string) error {

	if username == nil {
		return errors.New("no username provided")
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
				primitive.E{Key: "counters.votes.comments", Value: -1},
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

func (repo *Repository) Close() error {
	ctx, cancel := context.WithTimeout(context.Background(), repo.cfg.RequestTimeout)
	defer cancel()
	return repo.client.Disconnect(ctx)
}
