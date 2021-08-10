package api

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
	"github.com/rgynn/ptrconv"
)

func (svc *Service) SearchUsersHandler(c echo.Context) error {

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

	result, err := svc.users.Search(ctx, username, ptrconv.StringPtr("user"), from, size)
	if err != nil {
		c.Logger().Warnf("Failed to search for users: %s", err.Error())
		return errors.New("error occured")
	}

	return c.JSON(http.StatusOK, result)
}
