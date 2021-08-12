package api

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/rgynn/ptrconv"
)

func (svc *Service) SearchUsersHandler(w http.ResponseWriter, r *http.Request) {

	ctx := r.Context()

	var username *string
	if uname := r.URL.Query().Get("username"); uname != "" {
		username = &uname
	}

	from, err := strconv.ParseInt(r.URL.Query().Get("from"), 10, 64)
	if err != nil {
		from = 0
	}

	size, err := strconv.ParseInt(r.URL.Query().Get("size"), 10, 64)
	if err != nil || size < 1 {
		size = 100
	}

	claims, err := ClaimsFromContext(ctx)
	if err != nil {
		NewErrorResponse(w, r, http.StatusUnauthorized, err)
		return
	}

	if claims.IsUser() {
		NewErrorResponse(w, r, http.StatusUnauthorized, err)
		return
	}

	result, err := svc.users.Search(ctx, username, ptrconv.StringPtr("user"), from, size)
	if err != nil {
		NewErrorResponse(w, r, http.StatusInternalServerError, err)
		return
	}

	if err := svc.MarshalJSONResponse(w, http.StatusOK, result); err != nil {
		NewErrorResponse(w, r, http.StatusInternalServerError, err)
		return
	}
}

func (svc *Service) DeactivateUserHandler(w http.ResponseWriter, r *http.Request) {

	username := mux.Vars(r)["username"]
	ctx := r.Context()

	claims, err := ClaimsFromContext(ctx)
	if err != nil {
		NewErrorResponse(w, r, http.StatusUnauthorized, err)
		return
	}

	if claims.Username == nil || *claims.Username != username {
		NewErrorResponse(w, r, http.StatusUnauthorized, errors.New("cannot deactivate another account"))
		return
	}

	if err := svc.users.Deactivate(ctx, claims.Username, claims.Role); err != nil {
		NewErrorResponse(w, r, http.StatusInternalServerError, err)
		return
	}

	if err := svc.NoContentResponse(w, http.StatusAccepted); err != nil {
		NewErrorResponse(w, r, http.StatusInternalServerError, err)
		return
	}
}
