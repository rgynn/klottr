package mongo

import (
	"context"
	"errors"

	"github.com/rgynn/klottr/pkg/comment"
	"github.com/rgynn/klottr/pkg/config"
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
	return errors.New("not implemented yet")
}

func (repo *Repository) Get(ctx context.Context, id *string) (*comment.Model, error) {
	return nil, errors.New("not implemented yet")
}

func (repo *Repository) ListByThreadID(ctx context.Context, threadID *primitive.ObjectID, from, size int64) ([]*comment.Model, error) {
	return nil, errors.New("not implemented yet")
}

func (repo *Repository) ListByUserID(ctx context.Context, userID *primitive.ObjectID, from, size int64) ([]*comment.Model, error) {
	return nil, errors.New("not implemented yet")
}

func (repo *Repository) Delete(ctx context.Context, id *string) error {
	return errors.New("not implemented yet")
}

func (repo *Repository) IncVotes(ctx context.Context, id *string) error {
	return errors.New("not implemented yet")
}

func (repo *Repository) DecVotes(ctx context.Context, id *string) error {
	return errors.New("not implemented yet")
}

func (repo *Repository) Close() error {
	ctx, cancel := context.WithTimeout(context.Background(), repo.cfg.RequestTimeout)
	defer cancel()
	return repo.client.Disconnect(ctx)
}
