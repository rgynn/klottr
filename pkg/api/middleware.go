package api

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt"
	"github.com/rgynn/klottr/pkg/helper"
	"github.com/rgynn/klottr/pkg/user"
	"github.com/rgynn/ptrconv"
	"github.com/sirupsen/logrus"
)

// Request ID

type RequestIDContextKey struct{}

func RequestIDContext(ctx context.Context, rid string) context.Context {
	return context.WithValue(ctx, RequestIDContextKey{}, rid)
}

func RequestIDFromContext(ctx context.Context) (*string, error) {
	rid, ok := ctx.Value(RequestIDContextKey{}).(string)
	if !ok {
		return nil, errors.New("failed to type assert request id (string) from context")
	}
	return &rid, nil
}

func (svc *Service) RequestIDMiddleware(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reqid := r.Header.Get("X-Request-ID")
		if reqid == "" {
			reqid = helper.RandomString(20)
		}
		h.ServeHTTP(w, r.WithContext(RequestIDContext(r.Context(), reqid)))
	})
}

// Context Logger

type LoggerContextKey struct{}

func LoggerContext(ctx context.Context, contextlogger *logrus.Entry) context.Context {
	return context.WithValue(ctx, LoggerContextKey{}, contextlogger)
}

func LoggerFromContext(ctx context.Context) (*logrus.Entry, error) {
	contextlogger, ok := ctx.Value(LoggerContextKey{}).(*logrus.Entry)
	if !ok {
		return nil, errors.New("failed to type assert *logrus.Logger from context")
	}
	return contextlogger, nil
}

func (svc *Service) ContextLoggerMiddleware(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		contextlogger := logrus.New().WithFields(logrus.Fields{
			"start":  time.Now().UTC().Format(time.RFC3339),
			"method": r.Method,
			"path":   r.URL.Path,
			"query":  r.URL.Query().Encode(),
		})
		if rid, err := RequestIDFromContext(r.Context()); err == nil {
			contextlogger = contextlogger.WithField("rid", *rid)
		}
		h.ServeHTTP(w, r.WithContext(LoggerContext(r.Context(), contextlogger)))
	})
}

func (svc *Service) AccessLoggerMiddleware(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h.ServeHTTP(w, r)
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
