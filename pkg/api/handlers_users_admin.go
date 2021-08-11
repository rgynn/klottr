package api

import (
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/rgynn/klottr/pkg/thread"
	"github.com/rgynn/klottr/pkg/user"
	"github.com/rgynn/ptrconv"
)

func (svc *Service) CreateAdminUserHandler(w http.ResponseWriter, r *http.Request) {

	m := new(user.Model)
	ctx := r.Context()

	claims, err := ClaimsFromContext(ctx)
	if err != nil {
		NewErrorResponse(w, r, http.StatusUnauthorized, err)
		return
	}

	if claims.IsUser() {
		NewErrorResponse(w, r, http.StatusUnauthorized, err)
		return
	}

	if err := svc.UnmarshalJSONRequest(w, r, &m); err != nil {
		NewErrorResponse(w, r, http.StatusBadRequest, err)
		return
	}

	m.Role = ptrconv.StringPtr("admin")

	if err := m.HashPassword(); err != nil {
		NewErrorResponse(w, r, http.StatusBadRequest, err)
		return
	}

	if err := m.HashEmail(); err != nil {
		NewErrorResponse(w, r, http.StatusBadRequest, err)
		return
	}

	if err := m.ValidForSave(); err != nil {
		NewErrorResponse(w, r, http.StatusBadRequest, err)
		return
	}

	if err := svc.users.Create(ctx, m); err != nil {
		NewErrorResponse(w, r, http.StatusInternalServerError, err)
		return
	}

	if err := svc.NoContentResponse(w, http.StatusCreated); err != nil {
		NewErrorResponse(w, r, http.StatusInternalServerError, err)
		return
	}
}

func (svc *Service) SearchAdminUsersHandler(w http.ResponseWriter, r *http.Request) {

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

	result, err := svc.users.Search(ctx, username, ptrconv.StringPtr("admin"), from, size)
	if err != nil {
		NewErrorResponse(w, r, http.StatusInternalServerError, thread.ErrCategoryNotFound)
		return
	}

	if err := svc.MarshalJSONResponse(w, http.StatusOK, result); err != nil {
		NewErrorResponse(w, r, http.StatusInternalServerError, err)
		return
	}
}

func (svc *Service) GetAdminUserHandler(w http.ResponseWriter, r *http.Request) {

	username := mux.Vars(r)["username"]
	ctx := r.Context()

	claims, err := ClaimsFromContext(ctx)
	if err != nil {
		NewErrorResponse(w, r, http.StatusUnauthorized, err)
		return
	}

	if claims.IsUser() {
		NewErrorResponse(w, r, http.StatusUnauthorized, err)
		return
	}

	if username != ptrconv.StringPtrString(claims.Username) {
		NewErrorResponse(w, r, http.StatusUnauthorized, err)
		return
	}

	result, err := svc.users.Get(ctx, &username)
	if err != nil {
		NewErrorResponse(w, r, http.StatusInternalServerError, err)
		return
	}

	if err := svc.MarshalJSONResponse(w, http.StatusOK, result); err != nil {
		NewErrorResponse(w, r, http.StatusInternalServerError, err)
		return
	}
}

func (svc *Service) DeleteAdminUserHandler(w http.ResponseWriter, r *http.Request) {

	username := mux.Vars(r)["username"]
	ctx := r.Context()

	claims, err := ClaimsFromContext(ctx)
	if err != nil {
		NewErrorResponse(w, r, http.StatusUnauthorized, err)
		return
	}

	if claims.IsUser() {
		NewErrorResponse(w, r, http.StatusUnauthorized, err)
		return
	}

	if username != ptrconv.StringPtrString(claims.Username) {
		NewErrorResponse(w, r, http.StatusUnauthorized, err)
		return
	}

	if err := svc.users.Delete(ctx, &username, ptrconv.StringPtr("admin")); err != nil {
		NewErrorResponse(w, r, http.StatusInternalServerError, err)
		return
	}

	if err := svc.NoContentResponse(w, http.StatusAccepted); err != nil {
		NewErrorResponse(w, r, http.StatusInternalServerError, err)
		return
	}
}
