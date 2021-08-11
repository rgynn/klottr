package api

import (
	"context"
	"errors"
	"math/rand"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt"
	"github.com/rgynn/klottr/pkg/user"
	"github.com/rgynn/ptrconv"
	"github.com/sirupsen/logrus"
)

// Request ID

type RequestIDContextKey struct{}

func RequestIDContext(ctx context.Context, reqid string) context.Context {
	return context.WithValue(ctx, RequestIDContextKey{}, reqid)
}

func RequestIDFromContext(ctx context.Context) (*string, error) {
	reqid, ok := ctx.Value(RequestIDContextKey{}).(string)
	if !ok {
		return nil, errors.New("failed to type assert request id (string) from context")
	}
	return &reqid, nil
}

func randomString(n int) string {
	var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")
	s := make([]rune, n)
	for i := range s {
		s[i] = letters[rand.Intn(len(letters))]
	}
	return string(s)
}

func (svc *Service) RequestIDMiddleware(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h.ServeHTTP(w, r.WithContext(RequestIDContext(r.Context(), randomString(20))))
	})
}

// Context Logger

type LoggerContextKey struct{}

func LoggerContext(ctx context.Context, contextlogger *logrus.Logger) context.Context {
	return context.WithValue(ctx, LoggerContextKey{}, contextlogger)
}

func LoggerFromContext(ctx context.Context) (*logrus.Logger, error) {
	contextlogger, ok := ctx.Value(LoggerContextKey{}).(*logrus.Logger)
	if !ok {
		return nil, errors.New("failed to type assert *logrus.Logger from context")
	}
	return contextlogger, nil
}

func (svc *Service) ContextLoggerMiddleware(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h.ServeHTTP(w, r.WithContext(LoggerContext(r.Context(), logrus.New())))
	})
}

// JWT

type ClaimsContextKey struct{}

func ClaimsContext(ctx context.Context, claims *JWTClaims) context.Context {
	return context.WithValue(ctx, ClaimsContextKey{}, claims)
}

func ClaimsFromContext(ctx context.Context) (*JWTClaims, error) {
	claims, ok := ctx.Value(ClaimsContextKey{}).(*JWTClaims)
	if !ok {
		return nil, errors.New("failed to type assert *JWTClaims from context")
	}
	return claims, nil
}

func (svc *Service) JWTMiddleware(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		tokenString := strings.Replace(r.Header.Get("Authorization"), "Bearer ", "", 1)
		if tokenString == "" {
			NewErrorResponse(w, r, http.StatusUnauthorized, errors.New("no valid jwt provided"))
			return
		}

		claims := new(JWTClaims)

		_, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
			return []byte(svc.cfg.JWTSecret), nil
		})
		if err != nil {
			NewErrorResponse(w, r, http.StatusUnauthorized, errors.New("no valid jwt provided"))
			return
		}

		h.ServeHTTP(w, r.WithContext(ClaimsContext(r.Context(), claims)))
	})
}

type JWTClaims struct {
	Username  *string       `json:"username"`
	UserID    *string       `json:"userID"`
	Role      *string       `json:"role"`
	Validated bool          `json:"validated"`
	Counters  user.Counters `json:"counters"`
	jwt.StandardClaims
}

func (claims *JWTClaims) IsAdmin() bool {
	return ptrconv.StringPtrString(claims.Role) == "admin"
}

func (claims *JWTClaims) IsUser() bool {
	return ptrconv.StringPtrString(claims.Role) == "user"
}
