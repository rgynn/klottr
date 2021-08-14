package api

import (
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/rgynn/klottr/pkg/comment"
	"github.com/rgynn/klottr/pkg/thread"
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

	if err := svc.UnmarshalJSONRequest(w, r, &m); err != nil {
		NewErrorResponse(w, r, http.StatusBadRequest, err)
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
		NewErrorResponse(w, r, http.StatusInternalServerError, err)
		return
	}

	switch category {
	case "misc":
		if err := svc.misc.IncComments(ctx, &slugID, &slugTitle); err != nil {
			NewErrorResponse(w, r, http.StatusInternalServerError, err)
			return
		}
	}

	if err := svc.users.IncCommentsCounter(ctx, claims.Username); err != nil {
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

	_, err := ClaimsFromContext(ctx)
	if err != nil {
		NewErrorResponse(w, r, http.StatusUnauthorized, err)
		return
	}

	switch category {
	case "misc":
		if _, err := svc.misc.Get(ctx, &slugID, &slugTitle); err != nil {
			NewErrorResponse(w, r, http.StatusInternalServerError, err)
			return
		}
	default:
		NewErrorResponse(w, r, http.StatusInternalServerError, thread.ErrCategoryNotFound)
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

	switch category {
	case "misc":
		if _, err := svc.misc.Get(ctx, &slugID, &slugTitle); err != nil {
			NewErrorResponse(w, r, http.StatusInternalServerError, err)
			return
		}
	default:
		NewErrorResponse(w, r, http.StatusNotFound, thread.ErrCategoryNotFound)
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
		NewErrorResponse(w, r, http.StatusInternalServerError, err)
		return
	}

	if err := svc.users.DecCommentsCounter(ctx, claims.UserID); err != nil {
		NewErrorResponse(w, r, http.StatusInternalServerError, err)
		return
	}

	if err := svc.NoContentResponse(w, http.StatusAccepted); err != nil {
		NewErrorResponse(w, r, http.StatusInternalServerError, err)
		return
	}
}

func (svc *Service) UpVoteCommentHandler(w http.ResponseWriter, r *http.Request) {

	vars := mux.Vars(r)
	category := vars["category"]
	slugID := vars["slug_id"]
	slugTitle := vars["slug_title"]
	commentSlugID := vars["comment_slug_id"]
	ctx := r.Context()

	_, err := ClaimsFromContext(ctx)
	if err != nil {
		NewErrorResponse(w, r, http.StatusUnauthorized, err)
		return
	}

	switch category {
	case "misc":
		if _, err := svc.misc.Get(ctx, &slugID, &slugTitle); err != nil {
			NewErrorResponse(w, r, http.StatusInternalServerError, err)
			return
		}
	default:
		NewErrorResponse(w, r, http.StatusNotFound, thread.ErrCategoryNotFound)
		return
	}

	cmnt, err := svc.comments.Get(ctx, &commentSlugID)
	if err != nil {
		NewErrorResponse(w, r, http.StatusNotFound, thread.ErrCategoryNotFound)
		return
	}

	if err := svc.comments.IncVotes(ctx, &commentSlugID); err != nil {
		NewErrorResponse(w, r, http.StatusInternalServerError, err)
		return
	}

	if err := svc.users.IncCommentsVotes(ctx, cmnt.Username); err != nil {
		NewErrorResponse(w, r, http.StatusInternalServerError, err)
		return
	}

	if err := svc.NoContentResponse(w, http.StatusAccepted); err != nil {
		NewErrorResponse(w, r, http.StatusInternalServerError, err)
		return
	}
}

func (svc *Service) DownVoteCommentHandler(w http.ResponseWriter, r *http.Request) {

	vars := mux.Vars(r)
	category := vars["category"]
	slugID := vars["slug_id"]
	slugTitle := vars["slug_title"]
	commentSlugID := vars["comment_slug_id"]
	ctx := r.Context()

	_, err := ClaimsFromContext(ctx)
	if err != nil {
		NewErrorResponse(w, r, http.StatusUnauthorized, err)
		return
	}

	switch category {
	case "misc":
		if _, err := svc.misc.Get(ctx, &slugID, &slugTitle); err != nil {
			NewErrorResponse(w, r, http.StatusInternalServerError, err)
			return
		}
	default:
		NewErrorResponse(w, r, http.StatusNotFound, thread.ErrCategoryNotFound)
		return
	}

	cmnt, err := svc.comments.Get(ctx, &commentSlugID)
	if err != nil {
		NewErrorResponse(w, r, http.StatusNotFound, thread.ErrCategoryNotFound)
		return
	}

	if err := svc.comments.DecVotes(ctx, &commentSlugID); err != nil {
		NewErrorResponse(w, r, http.StatusInternalServerError, err)
		return
	}

	if err := svc.users.DecCommentsVotes(ctx, cmnt.Username); err != nil {
		NewErrorResponse(w, r, http.StatusInternalServerError, err)
		return
	}

	if err := svc.NoContentResponse(w, http.StatusAccepted); err != nil {
		NewErrorResponse(w, r, http.StatusInternalServerError, err)
		return
	}
}
