package api

import (
	"net/http"
	"strconv"

	"github.com/rgynn/ptrconv"
)

func (svc *Service) SearchUsersHandler(w http.ResponseWriter, r *http.Request) {

	ctx := r.Context()

	claims, err := ClaimsFromContext(ctx)
	if err != nil {
		NewErrorResponse(w, r, http.StatusUnauthorized, err)
		return
	}

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
