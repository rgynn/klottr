package api

import (
	"errors"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt"
	"github.com/labstack/echo/v4"
	"github.com/rgynn/klottr/pkg/user"
	"github.com/rgynn/ptrconv"
)

type LoginInput struct {
	Username *string `json:"username,omitempty"`
	Password *string `json:"password,omitempty"`
}

func (input *LoginInput) Valid() error {

	if input == nil {
		return errors.New("no login input provided")
	}

	if input.Username == nil {
		return errors.New("no username provided")
	}

	if input.Password == nil {
		return errors.New("no password provided")
	}

	return nil
}

type JWTClaims struct {
	Username  *string       `json:"username"`
	UserID    *string       `json:"userID"`
	Role      *string       `json:"role"`
	Validated bool          `json:"validated"`
	Counters  user.Counters `json:"counters"`
	jwt.StandardClaims
}

func (claims *JWTClaims) IsAdmin() bool {
	return ptrconv.StringPtrString(claims.Role) == "admin"
}

func (claims *JWTClaims) IsUser() bool {
	return ptrconv.StringPtrString(claims.Role) == "user"
}

func (svc *Service) GetClaims(c echo.Context) (*JWTClaims, error) {

	user, ok := c.Get("user").(*jwt.Token)
	if !ok {
		return nil, errors.New("failed to type assert *jwt.Token from context")
	}

	claims, ok := user.Claims.(*JWTClaims)
	if !ok {
		return nil, errors.New("failed to type assert JWTClaims from context")
	}

	return claims, nil
}

func (svc *Service) SignInHandler(c echo.Context) error {

	m := new(LoginInput)
	ctx := c.Request().Context()

	if err := c.Bind(m); err != nil {
		return errors.New("error occured")
	}

	if err := m.Valid(); err != nil {
		return echo.ErrBadRequest
	}

	u, err := svc.users.Get(ctx, m.Username)
	if err != nil {
		c.Logger().Warnf("Failed to find user: %s", err.Error())
		return errors.New("error occured")
	}

	if err := u.ValidPassword(m.Password); err != nil {
		return echo.ErrUnauthorized
	}

	claims := &JWTClaims{
		Username:  u.Username,
		UserID:    ptrconv.StringPtr(u.ID.String()),
		Validated: u.Validated,
		Counters:  u.Counters,
		Role:      u.Role,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(time.Hour * 72).Unix(),
		},
	}

	t, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(svc.cfg.JWTSecret))
	if err != nil {
		return echo.ErrUnauthorized
	}

	return c.JSON(http.StatusOK, echo.Map{
		"token": t,
	})
}

func (svc *Service) SignUpHandler(c echo.Context) error {

	m := new(user.Model)
	ctx := c.Request().Context()

	if err := c.Bind(m); err != nil {
		return echo.ErrBadRequest
	}

	m.Role = ptrconv.StringPtr("user")

	if err := m.HashPassword(); err != nil {
		return echo.ErrBadRequest
	}

	if err := m.HashEmail(); err != nil {
		return echo.ErrBadRequest
	}

	if err := m.ValidForSave(); err != nil {
		return echo.ErrBadRequest
	}

	if err := svc.users.Create(ctx, m); err != nil {
		c.Logger().Warnf("Failed to create user in users repository: %s", err.Error())
		return errors.New("error occured")
	}

	return c.NoContent(http.StatusCreated)
}

func (svc *Service) DeactivateHandler(c echo.Context) error {

	username := c.Param("username")
	ctx := c.Request().Context()

	claims, err := svc.GetClaims(c)
	if err != nil {
		return echo.ErrUnauthorized
	}

	if claims.Username == nil || *claims.Username != username {
		return echo.ErrUnauthorized
	}

	if err := svc.users.Delete(ctx, claims.Username, claims.Role); err != nil {
		c.Logger().Warnf("Failed to delete user with username: %s, error: %s", *claims.Username, err.Error())
		return errors.New("error occured")
	}

	return c.NoContent(http.StatusAccepted)
}
