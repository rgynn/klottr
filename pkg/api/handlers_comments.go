package api

import (
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/rgynn/klottr/pkg/comment"
	"github.com/rgynn/klottr/pkg/thread"
	"github.com/rgynn/klottr/pkg/user"
	"github.com/rgynn/ptrconv"
)

func (svc *Service) CreateCommentHandler(w http.ResponseWriter, r *http.Request) {

	m := new(comment.Model)
	vars := mux.Vars(r)
	category := vars["category"]
	slugID := vars["slug_id"]
	slugTitle := vars["slug_title"]
	ctx := r.Context()

	claims, err := ClaimsFromContext(ctx)
	if err != nil {
		NewErrorResponse(w, r, http.StatusUnauthorized, err)
		return
	}

	logger, err := LoggerFromContext(ctx)
	if err != nil {
		NewErrorResponse(w, r, http.StatusInternalServerError, err)
		return
	}

	var thrd *thread.Model

	switch category {
	case "misc":
		thrd, err = svc.misc.Get(ctx, &slugID, &slugTitle)
	default:
		NewErrorResponse(w, r, http.StatusInternalServerError, thread.ErrCategoryNotFound)
		return
	}
	if err != nil {
		NewErrorResponse(w, r, http.StatusInternalServerError, err)
		return
	}

	if err := svc.UnmarshalJSONRequest(w, r, &m); err != nil {
		NewErrorResponse(w, r, http.StatusBadRequest, err)
		return
	}

	m.ThreadID = thrd.ID
	m.Username = claims.Username
	m.Created = *ptrconv.TimePtr(time.Now().UTC())

	if err := m.GenerateSlugs(); err != nil {
		NewErrorResponse(w, r, http.StatusInternalServerError, err)
		return
	}

	if err := m.ValidForSave(); err != nil {
		NewErrorResponse(w, r, http.StatusBadRequest, err)
		return
	}

	if err := svc.comments.Create(ctx, m); err != nil {
		logger.Warnf("failed to create user comment: %s", err.Error())
		NewErrorResponse(w, r, http.StatusInternalServerError, err)
		return
	}

	switch category {
	case "misc":
		err = svc.misc.IncCounter(ctx, &slugID, &slugTitle, ptrconv.StringPtr("counters.comments"), 1)
	}
	if err != nil {
		logger.Warnf("failed to increment %s thread num comments: %s", category, err.Error())
		NewErrorResponse(w, r, http.StatusInternalServerError, err)
		return
	}

	if err := svc.users.IncCounter(ctx, claims.Username, ptrconv.StringPtr("counters.num.comments"), 1); err != nil {
		logger.Warnf("failed to increment user num comments: %s", err.Error())
		NewErrorResponse(w, r, http.StatusInternalServerError, err)
		return
	}

	result, err := svc.comments.Get(ctx, m.SlugID)
	if err != nil {
		NewErrorResponse(w, r, http.StatusInternalServerError, err)
		return
	}

	if err := svc.MarshalJSONResponse(w, http.StatusCreated, result); err != nil {
		NewErrorResponse(w, r, http.StatusInternalServerError, err)
		return
	}
}

func (svc *Service) GetCommentHandler(w http.ResponseWriter, r *http.Request) {

	vars := mux.Vars(r)
	category := vars["category"]
	slugID := vars["slug_id"]
	slugTitle := vars["slug_title"]
	commentSlugID := vars["comment_slug_id"]
	ctx := r.Context()

	var err error

	switch category {
	case "misc":
		_, err = svc.misc.Get(ctx, &slugID, &slugTitle)
	default:
		NewErrorResponse(w, r, http.StatusInternalServerError, thread.ErrCategoryNotFound)
		return
	}
	if err != nil {
		NewErrorResponse(w, r, http.StatusInternalServerError, err)
		return
	}

	result, err := svc.comments.Get(ctx, &commentSlugID)
	if err != nil {
		NewErrorResponse(w, r, http.StatusInternalServerError, err)
		return
	}

	if err := svc.MarshalJSONResponse(w, http.StatusOK, result); err != nil {
		NewErrorResponse(w, r, http.StatusInternalServerError, err)
		return
	}
}

func (svc *Service) DeleteCommentHandler(w http.ResponseWriter, r *http.Request) {

	vars := mux.Vars(r)
	category := vars["category"]
	slugID := vars["slug_id"]
	slugTitle := vars["slug_title"]
	commentSlugID := vars["comment_slug_id"]
	ctx := r.Context()

	claims, err := ClaimsFromContext(ctx)
	if err != nil {
		NewErrorResponse(w, r, http.StatusUnauthorized, err)
		return
	}

	logger, err := LoggerFromContext(ctx)
	if err != nil {
		NewErrorResponse(w, r, http.StatusInternalServerError, err)
		return
	}

	switch category {
	case "misc":
		_, err = svc.misc.Get(ctx, &slugID, &slugTitle)
	default:
		NewErrorResponse(w, r, http.StatusNotFound, thread.ErrCategoryNotFound)
		return
	}
	if err != nil {
		NewErrorResponse(w, r, http.StatusInternalServerError, err)
		return
	}

	cmnt, err := svc.comments.Get(ctx, &commentSlugID)
	if err != nil {
		NewErrorResponse(w, r, http.StatusInternalServerError, err)
		return
	}

	if cmnt.Username != nil || *cmnt.Username != *claims.Username {
		NewErrorResponse(w, r, http.StatusUnauthorized, err)
		return
	}

	if err := svc.comments.Delete(ctx, &commentSlugID); err != nil {
		logger.Warnf("failed to delete user comment: %s", err.Error())
		NewErrorResponse(w, r, http.StatusInternalServerError, err)
		return
	}

	if err := svc.users.IncCounter(ctx, claims.UserID, ptrconv.StringPtr("counters.num.comments"), -1); err != nil {
		logger.Warnf("failed to decrement user comment count: %s", err.Error())
		NewErrorResponse(w, r, http.StatusInternalServerError, err)
		return
	}

	if err := svc.NoContentResponse(w, http.StatusAccepted); err != nil {
		NewErrorResponse(w, r, http.StatusInternalServerError, err)
		return
	}
}

func (svc *Service) VoteCommentHandler(w http.ResponseWriter, r *http.Request) {

	vars := mux.Vars(r)
	category := vars["category"]
	slugID := vars["slug_id"]
	slugTitle := vars["slug_title"]
	commentSlugID := vars["comment_slug_id"]
	ctx := r.Context()

	m := new(user.Vote)

	claims, err := ClaimsFromContext(ctx)
	if err != nil {
		NewErrorResponse(w, r, http.StatusUnauthorized, err)
		return
	}

	logger, err := LoggerFromContext(ctx)
	if err != nil {
		NewErrorResponse(w, r, http.StatusInternalServerError, err)
		return
	}

	switch category {
	case "misc":
		_, err = svc.misc.Get(ctx, &slugID, &slugTitle)
	default:
		NewErrorResponse(w, r, http.StatusNotFound, thread.ErrCategoryNotFound)
		return
	}
	if err != nil {
		NewErrorResponse(w, r, http.StatusInternalServerError, err)
		return
	}

	cmnt, err := svc.comments.Get(ctx, &commentSlugID)
	if err != nil {

		NewErrorResponse(w, r, http.StatusNotFound, thread.ErrCategoryNotFound)
		return
	}

	if err := svc.UnmarshalJSONRequest(w, r, &m); err != nil {
		NewErrorResponse(w, r, http.StatusBadRequest, err)
		return
	}

	if err := m.ValidForSave(); err != nil {
		NewErrorResponse(w, r, http.StatusBadRequest, err)
		return
	}

	if err := svc.comments.IncVotes(ctx, &commentSlugID, *m.Value); err != nil {
		logger.Warnf("failed to increment comment votes thread: %s", err.Error())
		NewErrorResponse(w, r, http.StatusInternalServerError, err)
		return
	}

	if err := svc.users.UpsertVote(ctx, claims.Username, m); err != nil {
		logger.Warnf("failed to upsert user vote: %s", err.Error())
		NewErrorResponse(w, r, http.StatusInternalServerError, err)
		return
	}

	if err := svc.users.IncCounter(ctx, cmnt.Username, ptrconv.StringPtr("counters.votes.comments"), *m.Value); err != nil {
		logger.Warnf("failed to increment user comment votes: %s", err.Error())
		NewErrorResponse(w, r, http.StatusInternalServerError, err)
		return
	}

	if err := svc.NoContentResponse(w, http.StatusAccepted); err != nil {
		NewErrorResponse(w, r, http.StatusInternalServerError, err)
		return
	}
}
