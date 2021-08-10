package api

import (
	"errors"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/rgynn/klottr/pkg/comment"
	"github.com/rgynn/klottr/pkg/thread"
)

func (svc *Service) CreateCommentHandler(c echo.Context) error {

	m := new(comment.Model)
	category := c.Param("category")
	threadID := c.Param("thread_id")
	ctx := c.Request().Context()

	claims, err := svc.GetClaims(c)
	if err != nil {
		return echo.ErrUnauthorized
	}

	if err := c.Bind(m); err != nil {
		return echo.ErrBadRequest
	}

	if err := m.ValidForSave(); err != nil {
		return echo.ErrBadRequest
	}

	switch category {
	case "misc":
		if _, err := svc.misc.Get(ctx, &threadID); err != nil {
			return thread.ErrNotFound
		}
	default:
		return thread.ErrCategoryNotFound
	}

	if err := svc.comments.Create(ctx, m); err != nil {
		return errors.New("error occured")
	}

	switch category {
	case "misc":
		if err := svc.misc.IncComments(ctx, &threadID); err != nil {
			return errors.New("error occured")
		}
	}

	if err := svc.users.IncCommentsCounter(ctx, claims.Username); err != nil {
		return errors.New("error occured")
	}

	return c.NoContent(http.StatusCreated)
}

func (svc *Service) GetCommentHandler(c echo.Context) error {

	category := c.Param("category")
	threadID := c.Param("thread_id")
	commentID := c.Param("comment_id")
	ctx := c.Request().Context()

	_, err := svc.GetClaims(c)
	if err != nil {
		return echo.ErrUnauthorized
	}

	switch category {
	case "misc":
		if _, err := svc.misc.Get(ctx, &threadID); err != nil {
			return thread.ErrNotFound
		}
	default:
		return thread.ErrCategoryNotFound
	}

	result, err := svc.comments.Get(ctx, &commentID)
	if err != nil {
		return errors.New("error occured")
	}

	return c.JSON(http.StatusOK, result)
}

func (svc *Service) DeleteCommentHandler(c echo.Context) error {

	category := c.Param("category")
	threadID := c.Param("thread_id")
	commentID := c.Param("comment_id")
	ctx := c.Request().Context()

	claims, err := svc.GetClaims(c)
	if err != nil {
		return echo.ErrUnauthorized
	}

	switch category {
	case "misc":
		if _, err := svc.misc.Get(ctx, &threadID); err != nil {
			return thread.ErrNotFound
		}
	default:
		return thread.ErrCategoryNotFound
	}

	cmnt, err := svc.comments.Get(ctx, &commentID)
	if err != nil {
		return errors.New("error occured")
	}

	if cmnt.UserID.String() != *claims.UserID {
		return echo.ErrUnauthorized
	}

	if err := svc.comments.Delete(ctx, &commentID); err != nil {
		return errors.New("error occured")
	}

	if err := svc.users.DecCommentsCounter(ctx, claims.UserID); err != nil {
		return errors.New("error occured")
	}

	return c.NoContent(http.StatusAccepted)
}

func (svc *Service) UpVoteCommentHandler(c echo.Context) error {

	category := c.Param("category")
	threadID := c.Param("thread_id")
	commentID := c.Param("comment_id")
	ctx := c.Request().Context()

	claims, err := svc.GetClaims(c)
	if err != nil {
		return echo.ErrUnauthorized
	}

	switch category {
	case "misc":
		if _, err := svc.misc.Get(ctx, &threadID); err != nil {
			return thread.ErrNotFound
		}
	default:
		return thread.ErrCategoryNotFound
	}

	if err := svc.comments.IncVotes(ctx, &commentID); err != nil {
		return errors.New("error occured")
	}

	if err := svc.users.IncCommentsVotes(ctx, claims.Username); err != nil {
		return errors.New("error occured")
	}

	return c.NoContent(http.StatusAccepted)
}

func (svc *Service) DownVoteCommentHandler(c echo.Context) error {

	category := c.Param("category")
	threadID := c.Param("thread_id")
	commentID := c.Param("comment_id")
	ctx := c.Request().Context()

	claims, err := svc.GetClaims(c)
	if err != nil {
		return echo.ErrUnauthorized
	}

	switch category {
	case "misc":
		if _, err := svc.misc.Get(ctx, &threadID); err != nil {
			return thread.ErrNotFound
		}
	default:
		return thread.ErrCategoryNotFound
	}

	if err := svc.comments.DecVotes(ctx, &commentID); err != nil {
		return errors.New("error occured")
	}

	if err := svc.users.DecCommentsVotes(ctx, claims.Username); err != nil {
		return errors.New("error occured")
	}

	return c.NoContent(http.StatusAccepted)
}
