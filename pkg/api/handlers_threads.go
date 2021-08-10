package api

import (
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/rgynn/klottr/pkg/thread"
	"github.com/rgynn/ptrconv"
)

func (svc *Service) CreateCategoryThreadHandler(c echo.Context) error {

	m := new(thread.Model)
	ctx := c.Request().Context()

	_, err := svc.GetClaims(c)
	if err != nil {
		return echo.ErrUnauthorized
	}

	if err := c.Bind(m); err != nil {
		return echo.ErrBadRequest
	}

	if err := m.ValidForSave(); err != nil {
		return echo.ErrBadRequest
	}

	m.Created = ptrconv.TimePtr(time.Now().UTC())

	switch *m.Category {
	case "misc":
		if err := svc.misc.Create(ctx, m); err != nil {
			c.Logger().Errorf("Failed to create %s thread in repository: %s", *m.Category, err.Error())
			return errors.New("error occured")
		}
	default:
		return errors.New("error occured")
	}

	return c.NoContent(http.StatusCreated)
}

func (svc *Service) ListCategoryThreadsHandler(c echo.Context) error {

	category := c.Param("category")
	ctx := c.Request().Context()

	_, err := svc.GetClaims(c)
	if err != nil {
		return echo.ErrUnauthorized
	}

	from, err := strconv.ParseInt(c.QueryParam("from"), 10, 64)
	if err != nil {
		from = 0
	}

	size, err := strconv.ParseInt(c.QueryParam("size"), 10, 64)
	if err != nil || size < 1 {
		size = 100
	}

	switch category {
	case "misc":
		result, err := svc.misc.List(ctx, from, size)
		if err != nil {
			c.Logger().Errorf("Failed to list %s threads in repository: %s", category, err.Error())
			return errors.New("error occured")
		}
		return c.JSON(http.StatusOK, result)
	default:
		return c.JSON(http.StatusNotFound, echo.Map{"message": thread.ErrCategoryNotFound.Error()})
	}
}

func (svc *Service) GetCategoryThreadHandler(c echo.Context) error {

	category := c.Param("category")
	id := c.Param("thread_id")
	ctx := c.Request().Context()

	_, err := svc.GetClaims(c)
	if err != nil {
		return echo.ErrUnauthorized
	}

	switch category {
	case "misc":
		result, err := svc.misc.Get(ctx, &id)
		if err != nil {
			c.Logger().Warnf("Failed to get thread: %s in %s thread repository: %s", id, category, err.Error())
			return errors.New("error occured")
		}
		return c.JSON(http.StatusOK, result)
	default:
		return errors.New("error occured")
	}
}

func (svc *Service) UpVoteCategoryThreadHandler(c echo.Context) error {

	category := c.Param("category")
	id := c.Param("thread_id")
	ctx := c.Request().Context()

	_, err := svc.GetClaims(c)
	if err != nil {
		return echo.ErrUnauthorized
	}

	switch category {
	case "misc":
		if err := svc.misc.IncVote(ctx, &id); err != nil {
			c.Logger().Warnf("Failed to increment votes on thread: %s in %s thread repository: %s", id, category, err.Error())
			return errors.New("error occured")
		}
	}

	return c.NoContent(http.StatusAccepted)
}

func (svc *Service) DownVoteCategoryThreadHandler(c echo.Context) error {

	category := c.Param("category")
	id := c.Param("thread_id")
	ctx := c.Request().Context()

	_, err := svc.GetClaims(c)
	if err != nil {
		return echo.ErrUnauthorized
	}

	switch category {
	case "misc":
		if err := svc.misc.DecVote(ctx, &id); err != nil {
			c.Logger().Warnf("Failed to decrement votes on thread: %s in %s thread repository: %s", id, category, err.Error())
			return errors.New("error occured")
		}
	}

	return c.NoContent(http.StatusAccepted)
}
