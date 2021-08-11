package api

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/rgynn/klottr/pkg/config"

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
	users    user.Repository
	misc     thread.Repository
	comments comment.Repository
}

func NewAPIFromConfig(cfg *config.Config) (*Service, error) {

	users, err := mongouser.NewRepository(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize users repository: %w", err)
	}

	misc, err := mongothread.NewRepository(cfg, "misc")
	if err != nil {
		return nil, fmt.Errorf("failed to initialize misc threads repository: %w", err)
	}

	comments, err := mongocomments.NewRepository(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize comments repository: %w", err)
	}

	return &Service{
		cfg:      cfg,
		users:    users,
		misc:     misc,
		comments: comments,
	}, nil
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
