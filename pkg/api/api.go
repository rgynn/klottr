package api

import (
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/rgynn/klottr/pkg/comment"
	"github.com/rgynn/klottr/pkg/config"

	"github.com/rgynn/klottr/pkg/thread"
	mongothread "github.com/rgynn/klottr/pkg/thread/mongo"

	"github.com/rgynn/klottr/pkg/user"
	mongouser "github.com/rgynn/klottr/pkg/user/mongo"
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
		return nil, fmt.Errorf("failed to initialize misc thread repository: %w", err)
	}

	return &Service{
		cfg:   cfg,
		users: users,
		misc:  misc,
	}, nil
}
