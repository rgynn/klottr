package api

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/rgynn/klottr/pkg/comment"
	"github.com/rgynn/klottr/pkg/thread"
)

func (svc *Service) CreateCommentHandler(w http.ResponseWriter, r *http.Request) {

	m := new(comment.Model)
	category := mux.Vars(r)["category"]
	threadID := mux.Vars(r)["thread_id"]
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

	if err := m.ValidForSave(); err != nil {
		NewErrorResponse(w, r, http.StatusBadRequest, err)
		return
	}

	switch category {
	case "misc":
		if _, err := svc.misc.Get(ctx, &threadID); err != nil {
			NewErrorResponse(w, r, http.StatusInternalServerError, err)
			return
		}
	default:
		NewErrorResponse(w, r, http.StatusInternalServerError, thread.ErrCategoryNotFound)
		return
	}

	if err := svc.comments.Create(ctx, m); err != nil {
		NewErrorResponse(w, r, http.StatusInternalServerError, err)
		return
	}

	switch category {
	case "misc":
		if err := svc.misc.IncComments(ctx, &threadID); err != nil {
			NewErrorResponse(w, r, http.StatusInternalServerError, err)
			return
		}
	}

	if err := svc.users.IncCommentsCounter(ctx, claims.Username); err != nil {
		NewErrorResponse(w, r, http.StatusInternalServerError, err)
		return
	}

	if err := svc.NoContentResponse(w, http.StatusCreated); err != nil {
		NewErrorResponse(w, r, http.StatusInternalServerError, err)
		return
	}
}

func (svc *Service) GetCommentHandler(w http.ResponseWriter, r *http.Request) {

	category := mux.Vars(r)["category"]
	threadID := mux.Vars(r)["thread_id"]
	commentID := mux.Vars(r)["comment_id"]
	ctx := r.Context()

	_, err := ClaimsFromContext(ctx)
	if err != nil {
		NewErrorResponse(w, r, http.StatusUnauthorized, err)
		return
	}

	switch category {
	case "misc":
		if _, err := svc.misc.Get(ctx, &threadID); err != nil {
			NewErrorResponse(w, r, http.StatusInternalServerError, err)
			return
		}
	default:
		NewErrorResponse(w, r, http.StatusInternalServerError, thread.ErrCategoryNotFound)
		return
	}

	result, err := svc.comments.Get(ctx, &commentID)
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

	category := mux.Vars(r)["category"]
	threadID := mux.Vars(r)["thread_id"]
	commentID := mux.Vars(r)["comment_id"]
	ctx := r.Context()

	claims, err := ClaimsFromContext(ctx)
	if err != nil {
		NewErrorResponse(w, r, http.StatusUnauthorized, err)
		return
	}

	switch category {
	case "misc":
		if _, err := svc.misc.Get(ctx, &threadID); err != nil {
			NewErrorResponse(w, r, http.StatusInternalServerError, err)
			return
		}
	default:
		NewErrorResponse(w, r, http.StatusNotFound, thread.ErrCategoryNotFound)
		return
	}

	cmnt, err := svc.comments.Get(ctx, &commentID)
	if err != nil {
		NewErrorResponse(w, r, http.StatusInternalServerError, err)
		return
	}

	if cmnt.UserID.String() != *claims.UserID {
		NewErrorResponse(w, r, http.StatusUnauthorized, err)
		return
	}

	if err := svc.comments.Delete(ctx, &commentID); err != nil {
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

	category := mux.Vars(r)["category"]
	threadID := mux.Vars(r)["thread_id"]
	commentID := mux.Vars(r)["comment_id"]
	ctx := r.Context()

	claims, err := ClaimsFromContext(ctx)
	if err != nil {
		NewErrorResponse(w, r, http.StatusUnauthorized, err)
		return
	}

	switch category {
	case "misc":
		if _, err := svc.misc.Get(ctx, &threadID); err != nil {
			NewErrorResponse(w, r, http.StatusInternalServerError, err)
			return
		}
	default:
		NewErrorResponse(w, r, http.StatusNotFound, thread.ErrCategoryNotFound)
		return
	}

	if err := svc.comments.IncVotes(ctx, &commentID); err != nil {
		NewErrorResponse(w, r, http.StatusInternalServerError, err)
		return
	}

	if err := svc.users.IncCommentsVotes(ctx, claims.Username); err != nil {
		NewErrorResponse(w, r, http.StatusInternalServerError, err)
		return
	}

	if err := svc.NoContentResponse(w, http.StatusAccepted); err != nil {
		NewErrorResponse(w, r, http.StatusInternalServerError, err)
		return
	}
}

func (svc *Service) DownVoteCommentHandler(w http.ResponseWriter, r *http.Request) {

	category := mux.Vars(r)["category"]
	threadID := mux.Vars(r)["thread_id"]
	commentID := mux.Vars(r)["comment_id"]
	ctx := r.Context()

	claims, err := ClaimsFromContext(ctx)
	if err != nil {
		NewErrorResponse(w, r, http.StatusUnauthorized, err)
		return
	}

	switch category {
	case "misc":
		if _, err := svc.misc.Get(ctx, &threadID); err != nil {
			NewErrorResponse(w, r, http.StatusInternalServerError, err)
			return
		}
	default:
		NewErrorResponse(w, r, http.StatusNotFound, thread.ErrCategoryNotFound)
		return
	}

	if err := svc.comments.DecVotes(ctx, &commentID); err != nil {
		NewErrorResponse(w, r, http.StatusInternalServerError, err)
		return
	}

	if err := svc.users.DecCommentsVotes(ctx, claims.Username); err != nil {
		NewErrorResponse(w, r, http.StatusInternalServerError, err)
		return
	}

	if err := svc.NoContentResponse(w, http.StatusAccepted); err != nil {
		NewErrorResponse(w, r, http.StatusInternalServerError, err)
		return
	}
}
