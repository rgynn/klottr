package main

import (
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/rgynn/klottr/pkg/api"
	"github.com/rgynn/klottr/pkg/config"
	"github.com/sirupsen/logrus"
)

func main() {

	cfg, err := config.NewFromEnv()
	if err != nil {
		logrus.Fatal(err)
	}

	api, err := api.NewAPIFromConfig(cfg)
	if err != nil {
		logrus.Fatal(err)
	}
	defer func() {
		if err := api.Close(); err != nil {
			logrus.Fatal(err)
		}
	}()

	r := mux.NewRouter()

	r.Use(
		api.RequestIDMiddleware,
		api.ContextLoggerMiddleware,
		api.JWTMiddleware,
	)

	v1 := r.PathPrefix("/api/1.0").Subrouter()

	// Version
	v1.HandleFunc("/version", api.VersionHandler).Methods(http.MethodGet)
	v1.HandleFunc("/healthz", api.HealthHandler).Methods(http.MethodGet)

	// Auth
	v1.HandleFunc("/auth/signin", api.SignInHandler).Methods(http.MethodPost)
	v1.HandleFunc("/auth/signup", api.SignUpHandler).Methods(http.MethodPost)
	v1.HandleFunc("/auth/deactivate", api.DeactivateHandler).Methods(http.MethodPost)

	// Threads
	v1.HandleFunc("/c/{category}", api.CreateThreadHandler).Methods(http.MethodPost)
	v1.HandleFunc("/c/{category}", api.ListThreadsHandler).Methods(http.MethodGet)
	v1.HandleFunc("/c/{category}/t/{slug_id}/{slug_title}", api.GetThreadHandler).Methods(http.MethodGet)
	v1.HandleFunc("/c/{category}/t/{slug_id}/{slug_title}/vote", api.VoteThreadHandler).Methods(http.MethodPost)

	// Comments
	v1.HandleFunc("/c/{category}/t/{slug_id}/{slug_title}/comments", api.CreateCommentHandler).Methods(http.MethodPost)
	v1.HandleFunc("/c/{category}/t/{slug_id}/{slug_title}/comments/{comment_slug_id}", api.GetCommentHandler).Methods(http.MethodGet)
	v1.HandleFunc("/c/{category}/t/{slug_id}/{slug_title}/comments/{comment_slug_id}", api.DeleteCommentHandler).Methods(http.MethodDelete)
	v1.HandleFunc("/c/{category}/t/{slug_id}/{slug_title}/comments/{comment_slug_id}/vote", api.VoteCommentHandler).Methods(http.MethodPost)

	srv := &http.Server{
		IdleTimeout:  cfg.IdleTimeout,
		ReadTimeout:  cfg.ReadTimeout,
		WriteTimeout: cfg.WriteTimeout,
		Addr:         cfg.Addr,
		Handler:      r,
	}

	logrus.Infof("Runing version: %s, built: %s\n", cfg.Version, cfg.BuildDate)
	logrus.Infof("Listening on http://%s\n", cfg.Addr)
	if err := srv.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}
