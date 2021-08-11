package api

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"github.com/rgynn/klottr/pkg/thread"
	"github.com/rgynn/ptrconv"
)

func (svc *Service) CreateCategoryThreadHandler(w http.ResponseWriter, r *http.Request) {

	m := new(thread.Model)
	ctx := r.Context()

	logger, err := LoggerFromContext(ctx)
	if err != nil {
		NewErrorResponse(w, r, http.StatusInternalServerError, err)
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

	m.Created = ptrconv.TimePtr(time.Now().UTC())

	switch *m.Category {
	case "misc":
		if err := svc.misc.Create(ctx, m); err != nil {
			logger.Errorf("Failed to create misc thread: %s", err.Error())
			NewErrorResponse(w, r, http.StatusInternalServerError, err)
			return
		}
	default:
		NewErrorResponse(w, r, http.StatusNotFound, thread.ErrCategoryNotFound)
		return
	}

	if err := svc.NoContentResponse(w, http.StatusCreated); err != nil {
		NewErrorResponse(w, r, http.StatusInternalServerError, err)
		return
	}
}

func (svc *Service) ListCategoryThreadsHandler(w http.ResponseWriter, r *http.Request) {

	category := mux.Vars(r)["category"]
	ctx := r.Context()

	logger, err := LoggerFromContext(ctx)
	if err != nil {
		NewErrorResponse(w, r, http.StatusInternalServerError, err)
		return
	}

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
		if err != nil {
			logger.Errorf("Failed to list misc threads: %s", err.Error())
			NewErrorResponse(w, r, http.StatusInternalServerError, err)
			return
		}
	default:
		NewErrorResponse(w, r, http.StatusNotFound, thread.ErrCategoryNotFound)
		return
	}

	if err := svc.MarshalJSONResponse(w, http.StatusOK, &result); err != nil {
		NewErrorResponse(w, r, http.StatusInternalServerError, err)
		return
	}
}

func (svc *Service) GetCategoryThreadHandler(w http.ResponseWriter, r *http.Request) {

	vars := mux.Vars(r)
	category := vars["category"]
	id := vars["thread_id"]
	ctx := r.Context()

	logger, err := LoggerFromContext(ctx)
	if err != nil {
		NewErrorResponse(w, r, http.StatusInternalServerError, err)
		return
	}

	var result *thread.Model

	switch category {
	case "misc":
		result, err = svc.misc.Get(ctx, &id)
	default:
		NewErrorResponse(w, r, http.StatusNotFound, thread.ErrCategoryNotFound)
		return
	}

	if err != nil {
		logger.Errorf("Failed to get misc thread: %s", err.Error())
		NewErrorResponse(w, r, http.StatusInternalServerError, err)
		return
	}

	if err := svc.MarshalJSONResponse(w, http.StatusOK, &result); err != nil {
		NewErrorResponse(w, r, http.StatusInternalServerError, err)
		return
	}
}

func (svc *Service) UpVoteCategoryThreadHandler(w http.ResponseWriter, r *http.Request) {

	vars := mux.Vars(r)
	category := vars["category"]
	id := vars["thread_id"]
	ctx := r.Context()

	logger, err := LoggerFromContext(ctx)
	if err != nil {
		NewErrorResponse(w, r, http.StatusInternalServerError, err)
		return
	}

	switch category {
	case "misc":
		if err := svc.misc.IncVote(ctx, &id); err != nil {
			logger.Errorf("Failed to upvote misc thread: %s", err.Error())
			NewErrorResponse(w, r, http.StatusInternalServerError, err)
			return
		}
	}

	if err := svc.NoContentResponse(w, http.StatusAccepted); err != nil {
		NewErrorResponse(w, r, http.StatusInternalServerError, err)
		return
	}
}

func (svc *Service) DownVoteCategoryThreadHandler(w http.ResponseWriter, r *http.Request) {

	vars := mux.Vars(r)
	category := vars["category"]
	id := vars["thread_id"]
	ctx := r.Context()

	logger, err := LoggerFromContext(ctx)
	if err != nil {
		NewErrorResponse(w, r, http.StatusInternalServerError, err)
		return
	}

	switch category {
	case "misc":
		if err := svc.misc.DecVote(ctx, &id); err != nil {
			logger.Errorf("Failed to upvote misc thread: %s", err.Error())
			NewErrorResponse(w, r, http.StatusInternalServerError, err)
			return
		}
	default:
		NewErrorResponse(w, r, http.StatusNotFound, thread.ErrCategoryNotFound)
		return
	}

	if err := svc.NoContentResponse(w, http.StatusAccepted); err != nil {
		NewErrorResponse(w, r, http.StatusInternalServerError, err)
		return
	}
}
