package api

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
	"github.com/rgynn/klottr/pkg/user"
	"github.com/rgynn/ptrconv"
)

func (svc *Service) CreateAdminUserHandler(c echo.Context) error {

	m := new(user.Model)
	ctx := c.Request().Context()

	claims, err := svc.GetClaims(c)
	if err != nil {
		return echo.ErrUnauthorized
	}

	if claims.IsUser() {
		return echo.ErrUnauthorized
	}

	if err := c.Bind(m); err != nil {
		return echo.ErrBadRequest
	}

	m.Role = ptrconv.StringPtr("admin")

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

func (svc *Service) SearchAdminUsersHandler(c echo.Context) error {

	ctx := c.Request().Context()

	var username *string
	if uname := c.QueryParam("username"); uname != "" {
		username = &uname
	}

	from, err := strconv.ParseInt(c.QueryParam("from"), 10, 64)
	if err != nil {
		from = 0
	}

	size, err := strconv.ParseInt(c.QueryParam("size"), 10, 64)
	if err != nil || size < 1 {
		size = 100
	}

	claims, err := svc.GetClaims(c)
	if err != nil {
		return echo.ErrUnauthorized
	}

	if claims.IsUser() {
		return echo.ErrUnauthorized
	}

	result, err := svc.users.Search(ctx, username, ptrconv.StringPtr("admin"), from, size)
	if err != nil {
		c.Logger().Warnf("Failed to search for admin users in users repository: %s", err.Error())
		return errors.New("error occured")
	}

	return c.JSON(http.StatusOK, result)
}

func (svc *Service) GetAdminUserHandler(c echo.Context) error {

	username := c.Param("username")
	ctx := c.Request().Context()

	claims, err := svc.GetClaims(c)
	if err != nil {
		return echo.ErrUnauthorized
	}

	if claims.IsUser() {
		return echo.ErrUnauthorized
	}

	if username != ptrconv.StringPtrString(claims.Username) {
		return echo.ErrUnauthorized
	}

	result, err := svc.users.Get(ctx, &username)
	if err != nil {
		c.Logger().Warnf("Failed to find user %s in repository: %s", username, err.Error())
		return errors.New("error occured")
	}

	return c.JSON(http.StatusOK, result)
}

func (svc *Service) DeleteAdminUserHandler(c echo.Context) error {

	username := c.Param("username")
	ctx := c.Request().Context()

	claims, err := svc.GetClaims(c)
	if err != nil {
		return echo.ErrUnauthorized
	}

	if claims.IsUser() {
		return echo.ErrUnauthorized
	}

	if username != ptrconv.StringPtrString(claims.Username) {
		return echo.ErrUnauthorized
	}

	if err := svc.users.Delete(ctx, &username, ptrconv.StringPtr("admin")); err != nil {
		c.Logger().Warnf("Failed to delete user with username: %s, error: %s", username, err.Error())
		return errors.New("error occured")
	}

	return c.NoContent(http.StatusAccepted)
}
