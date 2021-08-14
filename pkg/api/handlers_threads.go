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

	category := mux.Vars(r)["category"]
	m := new(thread.Model)
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

	if err := svc.UnmarshalJSONRequest(w, r, &m); err != nil {
		NewErrorResponse(w, r, http.StatusBadRequest, err)
		return
	}

	if err := m.GenerateSlugs(); err != nil {
		NewErrorResponse(w, r, http.StatusBadRequest, err)
		return
	}

	m.Category = &category
	m.Created = ptrconv.TimePtr(time.Now().UTC())

	if err := m.ValidForSave(); err != nil {
		NewErrorResponse(w, r, http.StatusBadRequest, err)
		return
	}

	var result *thread.Model

	switch category {
	case "misc":
		if err := svc.misc.Create(ctx, m); err != nil {
			logger.Errorf("Failed to create misc thread: %s", err.Error())
			NewErrorResponse(w, r, http.StatusInternalServerError, err)
			return
		}
		if result, err = svc.misc.Get(ctx, m.SlugID, m.SlugTitle); err != nil {
			logger.Errorf("Failed to get misc thread: %s", err.Error())
			NewErrorResponse(w, r, http.StatusInternalServerError, err)
			return
		}
	default:
		NewErrorResponse(w, r, http.StatusNotFound, thread.ErrCategoryNotFound)
		return
	}

	if err := svc.users.IncThreadsCounter(ctx, claims.Username); err != nil {
		NewErrorResponse(w, r, http.StatusInternalServerError, err)
		return
	}

	if err := svc.MarshalJSONResponse(w, http.StatusCreated, result); err != nil {
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
	slugID := vars["slug_id"]
	slugTitle := vars["slug_title"]
	ctx := r.Context()

	logger, err := LoggerFromContext(ctx)
	if err != nil {
		NewErrorResponse(w, r, http.StatusInternalServerError, err)
		return
	}

	var result *thread.Model

	switch category {
	case "misc":
		result, err = svc.misc.Get(ctx, &slugID, &slugTitle)
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
	slugID := vars["slug_id"]
	slugTitle := vars["slug_title"]
	ctx := r.Context()

	_, err := ClaimsFromContext(ctx)
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
		if err != nil {
			NewErrorResponse(w, r, http.StatusNotFound, thread.ErrNotFound)
			return
		}
		if err := svc.misc.IncVote(ctx, &slugID, &slugTitle); err != nil {
			logger.Errorf("Failed to upvote misc thread: %s", err.Error())
			NewErrorResponse(w, r, http.StatusInternalServerError, err)
			return
		}
	default:
		NewErrorResponse(w, r, http.StatusNotFound, thread.ErrCategoryNotFound)
		return
	}

	if err := svc.users.IncThreadsVotes(ctx, thrd.Username); err != nil {
		NewErrorResponse(w, r, http.StatusInternalServerError, err)
		return
	}

	if err := svc.NoContentResponse(w, http.StatusAccepted); err != nil {
		NewErrorResponse(w, r, http.StatusInternalServerError, err)
		return
	}
}

func (svc *Service) DownVoteCategoryThreadHandler(w http.ResponseWriter, r *http.Request) {

	vars := mux.Vars(r)
	category := vars["category"]
	slugID := vars["slug_id"]
	slugTitle := vars["slug_title"]
	ctx := r.Context()

	_, err := ClaimsFromContext(ctx)
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
		if err != nil {
			NewErrorResponse(w, r, http.StatusNotFound, thread.ErrNotFound)
			return
		}
		if err := svc.misc.DecVote(ctx, &slugID, &slugTitle); err != nil {
			logger.Errorf("Failed to upvote misc thread: %s", err.Error())
			NewErrorResponse(w, r, http.StatusInternalServerError, err)
			return
		}
	default:
		NewErrorResponse(w, r, http.StatusNotFound, thread.ErrCategoryNotFound)
		return
	}

	if err := svc.users.DecThreadsVotes(ctx, thrd.Username); err != nil {
		NewErrorResponse(w, r, http.StatusInternalServerError, err)
		return
	}

	if err := svc.NoContentResponse(w, http.StatusAccepted); err != nil {
		NewErrorResponse(w, r, http.StatusInternalServerError, err)
		return
	}
}
