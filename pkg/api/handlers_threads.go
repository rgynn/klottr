package api

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"github.com/rgynn/klottr/pkg/thread"
	"github.com/rgynn/klottr/pkg/user"
	"github.com/rgynn/ptrconv"
)

func (svc *Service) CreateThreadHandler(w http.ResponseWriter, r *http.Request) {

	category := mux.Vars(r)["category"]
	ctx := r.Context()

	m := new(thread.Model)

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

	if err := svc.UnmarshalJSONRequest(w, r, &m); err != nil {
		NewErrorResponse(w, r, http.StatusBadRequest, err)
		return
	}

	if err := m.GenerateSlugs(); err != nil {
		NewErrorResponse(w, r, http.StatusBadRequest, err)
		return
	}

	m.Username = claims.Username
	m.Created = ptrconv.TimePtr(time.Now().UTC())

	if err := m.ValidForSave(); err != nil {
		NewErrorResponse(w, r, http.StatusBadRequest, err)
		return
	}

	var result *thread.Model

	switch category {
	case "misc":
		err = svc.misc.Create(ctx, m)
	default:
		NewErrorResponse(w, r, http.StatusNotFound, thread.ErrCategoryNotFound)
		return
	}
	if err != nil {
		logger.Errorf("Failed to create %s thread: %s", category, err.Error())
		NewErrorResponse(w, r, http.StatusInternalServerError, err)
		return
	}

	switch category {
	case "misc":
		result, err = svc.misc.Get(ctx, m.SlugID, m.SlugTitle)
	default:
		NewErrorResponse(w, r, http.StatusNotFound, thread.ErrCategoryNotFound)
		return
	}
	if err != nil {
		NewErrorResponse(w, r, http.StatusInternalServerError, err)
		return
	}

	if err := svc.users.IncCounter(ctx, claims.Username, ptrconv.StringPtr("counters.num.threads"), 1); err != nil {
		logger.Errorf("Failed to increment user num threads: %s", err.Error())
		NewErrorResponse(w, r, http.StatusInternalServerError, err)
		return
	}

	if err := svc.MarshalJSONResponse(w, http.StatusCreated, result); err != nil {
		NewErrorResponse(w, r, http.StatusInternalServerError, err)
		return
	}
}

func (svc *Service) ListThreadsHandler(w http.ResponseWriter, r *http.Request) {

	category := mux.Vars(r)["category"]
	ctx := r.Context()

	from, err := strconv.ParseInt(r.URL.Query().Get("from"), 10, 64)
	if err != nil {
		from = 0
	}

	size, err := strconv.ParseInt(r.URL.Query().Get("size"), 10, 64)
	if err != nil || size < 1 {
		size = 100
	}

	result := []*thread.Model{}

	switch category {
	case "misc":
		result, err = svc.misc.List(ctx, from, size)
	default:
		NewErrorResponse(w, r, http.StatusNotFound, thread.ErrCategoryNotFound)
		return
	}
	if err != nil {
		NewErrorResponse(w, r, http.StatusInternalServerError, err)
		return
	}

	if err := svc.MarshalJSONResponse(w, http.StatusOK, &result); err != nil {
		NewErrorResponse(w, r, http.StatusInternalServerError, err)
		return
	}
}

func (svc *Service) GetThreadHandler(w http.ResponseWriter, r *http.Request) {

	vars := mux.Vars(r)
	category := vars["category"]
	slugID := vars["slug_id"]
	slugTitle := vars["slug_title"]
	ctx := r.Context()

	var result *thread.Model
	var err error

	switch category {
	case "misc":
		result, err = svc.misc.Get(ctx, &slugID, &slugTitle)
	default:
		NewErrorResponse(w, r, http.StatusNotFound, thread.ErrCategoryNotFound)
		return
	}
	if err != nil {
		NewErrorResponse(w, r, http.StatusInternalServerError, err)
		return
	}

	if err := svc.MarshalJSONResponse(w, http.StatusOK, &result); err != nil {
		NewErrorResponse(w, r, http.StatusInternalServerError, err)
		return
	}
}

func (svc *Service) VoteThreadHandler(w http.ResponseWriter, r *http.Request) {

	vars := mux.Vars(r)
	category := vars["category"]
	slugID := vars["slug_id"]
	slugTitle := vars["slug_title"]
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

	var thrd *thread.Model

	switch category {
	case "misc":
		thrd, err = svc.misc.Get(ctx, &slugID, &slugTitle)
	default:
		NewErrorResponse(w, r, http.StatusNotFound, thread.ErrCategoryNotFound)
		return
	}
	if err != nil {
		NewErrorResponse(w, r, http.StatusNotFound, err)
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

	switch category {
	case "misc":
		err = svc.misc.IncCounter(ctx, &slugID, &slugTitle, ptrconv.StringPtr("counters.votes"), *m.Value)
	default:
		NewErrorResponse(w, r, http.StatusNotFound, thread.ErrCategoryNotFound)
		return
	}
	if err != nil {
		logger.Errorf("Failed to increment %s thread votes: %s", category, err.Error())
		NewErrorResponse(w, r, http.StatusInternalServerError, err)
		return
	}

	if err := svc.users.IncCounter(ctx, thrd.Username, ptrconv.StringPtr("counters.votes.threads"), *m.Value); err != nil {
		logger.Errorf("Failed to increment user thread votes: %s", err.Error())
		NewErrorResponse(w, r, http.StatusInternalServerError, err)
		return
	}

	if err := svc.users.UpsertVote(ctx, claims.Username, m); err != nil {
		logger.Errorf("Failed to upsert user vote: %s", err.Error())
		NewErrorResponse(w, r, http.StatusInternalServerError, err)
		return
	}

	if err := svc.NoContentResponse(w, http.StatusAccepted); err != nil {
		NewErrorResponse(w, r, http.StatusInternalServerError, err)
		return
	}
}
