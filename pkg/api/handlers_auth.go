package api

import (
	"errors"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt"
	"github.com/gorilla/mux"
	"github.com/rgynn/klottr/pkg/user"
	"github.com/rgynn/ptrconv"
)

type LoginInput struct {
	Username *string `json:"username,omitempty"`
	Password *string `json:"password,omitempty"`
}

func (input *LoginInput) Valid() error {

	if input == nil {
		return errors.New("no login input provided")
	}

	if input.Username == nil {
		return errors.New("no username provided")
	}

	if input.Password == nil {
		return errors.New("no password provided")
	}

	return nil
}

func (svc *Service) SignInHandler(w http.ResponseWriter, r *http.Request) {

	m := new(LoginInput)
	ctx := r.Context()

	if err := svc.UnmarshalJSONRequest(w, r, &m); err != nil {
		NewErrorResponse(w, r, http.StatusBadRequest, err)
		return
	}

	if err := m.Valid(); err != nil {
		NewErrorResponse(w, r, http.StatusBadRequest, err)
		return
	}

	u, err := svc.users.Get(ctx, m.Username)
	if err != nil {
		NewErrorResponse(w, r, http.StatusInternalServerError, err)
		return
	}

	if err := u.ValidPassword(m.Password); err != nil {
		NewErrorResponse(w, r, http.StatusUnauthorized, err)
		return
	}

	claims := &JWTClaims{
		Username:  u.Username,
		UserID:    ptrconv.StringPtr(u.ID.String()),
		Validated: u.Validated,
		Counters:  u.Counters,
		Role:      u.Role,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(time.Hour * 72).Unix(),
		},
	}

	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(svc.cfg.JWTSecret))
	if err != nil {
		NewErrorResponse(w, r, http.StatusUnauthorized, err)
		return
	}

	if err := svc.MarshalJSONResponse(w, http.StatusOK, map[string]string{"token": token}); err != nil {
		NewErrorResponse(w, r, http.StatusInternalServerError, err)
		return
	}
}

func (svc *Service) SignUpHandler(w http.ResponseWriter, r *http.Request) {

	m := new(user.Model)
	ctx := r.Context()

	if err := svc.UnmarshalJSONRequest(w, r, &m); err != nil {
		NewErrorResponse(w, r, http.StatusBadRequest, err)
		return
	}

	m.Role = ptrconv.StringPtr("user")

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

	if err := svc.NoContentResponse(w, http.StatusAccepted); err != nil {
		NewErrorResponse(w, r, http.StatusInternalServerError, err)
		return
	}
}

func (svc *Service) DeactivateHandler(w http.ResponseWriter, r *http.Request) {

	username := mux.Vars(r)["username"]
	ctx := r.Context()

	claims, err := ClaimsFromContext(ctx)
	if err != nil {
		NewErrorResponse(w, r, http.StatusUnauthorized, err)
		return
	}

	if claims.Username == nil || *claims.Username != username {
		NewErrorResponse(w, r, http.StatusUnauthorized, err)
		return
	}

	if err := svc.users.Delete(ctx, claims.Username, claims.Role); err != nil {
		NewErrorResponse(w, r, http.StatusInternalServerError, err)
		return
	}

	if err := svc.NoContentResponse(w, http.StatusAccepted); err != nil {
		NewErrorResponse(w, r, http.StatusInternalServerError, err)
		return
	}
}
