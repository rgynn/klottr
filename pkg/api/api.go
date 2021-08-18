package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/rgynn/klottr/pkg/config"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/rgynn/klottr/pkg/thread"
	mongothread "github.com/rgynn/klottr/pkg/thread/mongo"

	"github.com/rgynn/klottr/pkg/user"
	mongouser "github.com/rgynn/klottr/pkg/user/mongo"

	"github.com/rgynn/klottr/pkg/comment"
	mongocomments "github.com/rgynn/klottr/pkg/comment/mongo"
)

// BodyDumpFunc used to dump request and response bodies through logger
func (svc *Service) BodyDumpFunc(c echo.Context, reqBody, resBody []byte) {
	switch c.Request().Method {
	case http.MethodPost:
		c.Logger().Infof("REQUEST_BODY: %s", string(reqBody))
	}
}

// Service for api
type Service struct {
	cfg      *config.Config
	mongodb  *mongo.Client
	users    user.Repository
	misc     thread.Repository
	comments comment.Repository
	metrics  prometheus.Collector
}

func NewAPIFromConfig(cfg *config.Config) (*Service, error) {

	setupMetrics()

	mongodb, err := setupMongoDBConnection(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to setup connection to mongodb: %w", err)
	}

	users, err := mongouser.NewRepository(cfg, mongodb)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize users repository: %w", err)
	}

	misc, err := mongothread.NewRepository(cfg, mongodb, "misc")
	if err != nil {
		return nil, fmt.Errorf("failed to initialize misc threads repository: %w", err)
	}

	comments, err := mongocomments.NewRepository(cfg, mongodb)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize comments repository: %w", err)
	}

	return &Service{
		mongodb:  mongodb,
		cfg:      cfg,
		users:    users,
		misc:     misc,
		comments: comments,
	}, nil
}

func (svc *Service) Close() error {
	ctx, cancel := context.WithTimeout(context.Background(), svc.cfg.RequestTimeout)
	defer cancel()
	return svc.mongodb.Disconnect(ctx)
}

func (svc *Service) UnmarshalJSONRequest(w http.ResponseWriter, r *http.Request, v interface{}) error {
	switch r.Method {
	case http.MethodPost, http.MethodPut, http.MethodPatch:
		break
	default:
		return nil
	}
	readercloser := http.MaxBytesReader(w, r.Body, svc.cfg.RequestBodyLimitBytes)
	if err := json.NewDecoder(readercloser).Decode(&v); err != nil {
		return err
	}
	return nil
}

func (svc *Service) MarshalJSONResponse(w http.ResponseWriter, status int, v interface{}) error {
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(&v); err != nil {
		return err
	}
	return nil
}

func (svc *Service) NoContentResponse(w http.ResponseWriter, status int) error {
	w.WriteHeader(status)
	return nil
}

func setupMongoDBConnection(cfg *config.Config) (*mongo.Client, error) {

	ctx, cancel := context.WithTimeout(context.Background(), cfg.RequestTimeout)
	defer cancel()

	mongodb, err := mongo.Connect(ctx, options.Client().ApplyURI(cfg.DatabaseURL))
	if err != nil {
		return nil, err
	}

	if err := mongodb.Ping(ctx, nil); err != nil {
		return nil, err
	}

	return mongodb, nil
}
