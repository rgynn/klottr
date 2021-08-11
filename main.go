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
		log.Fatal(err)
	}

	api, err := api.NewAPIFromConfig(cfg)
	if err != nil {
		log.Fatal(err)
	}

	r := mux.NewRouter()

	r.Use(
		api.RequestIDMiddleware,
		api.ContextLoggerMiddleware,
	)

	r.HandleFunc("/auth/signin", api.SignInHandler).Methods(http.MethodPost)
	r.HandleFunc("/auth/signup", api.SignUpHandler).Methods(http.MethodPost)
	r.HandleFunc("/auth/deactivate", api.DeactivateHandler).Methods(http.MethodDelete)

	v1 := r.PathPrefix("/api/1.0").Subrouter()

	v1.Use(
		api.JWTMiddleware,
	)

	// Threads
	v1.HandleFunc("/c/{category}", api.CreateCategoryThreadHandler).Methods(http.MethodPost)
	v1.HandleFunc("/c/{category}", api.ListCategoryThreadsHandler).Methods(http.MethodGet)
	v1.HandleFunc("/c/{category}/t/{thread_id}", api.GetCategoryThreadHandler).Methods(http.MethodGet)
	v1.HandleFunc("/c/{category}/t/{thread_id}/upvote", api.UpVoteCategoryThreadHandler).Methods(http.MethodPost)
	v1.HandleFunc("/c/{category}/t/{thread_id}/downvote", api.DownVoteCategoryThreadHandler).Methods(http.MethodPost)

	// Comments
	v1.HandleFunc("/c/{category}/t/{thread_id}/com", api.CreateCommentHandler).Methods(http.MethodPost)
	v1.HandleFunc("/c/{category}/t/{thread_id}/com/{comment_id}", api.GetCommentHandler).Methods(http.MethodGet)
	v1.HandleFunc("/c/{category}/t/{thread_id}/com/{comment_id}", api.DeleteCommentHandler).Methods(http.MethodDelete)
	v1.HandleFunc("/c/{category}/t/{thread_id}/com/{comment_id}/upvote", api.UpVoteCommentHandler).Methods(http.MethodPost)
	v1.HandleFunc("/c/{category}/t/{thread_id}/com/{comment_id}/downvote", api.DownVoteCommentHandler).Methods(http.MethodPost)

	srv := &http.Server{
		IdleTimeout:  cfg.IdleTimeout,
		ReadTimeout:  cfg.ReadTimeout,
		WriteTimeout: cfg.WriteTimeout,
		Addr:         cfg.Addr,
		Handler:      r,
	}

	logrus.Infof("Listening on http://%s\n", cfg.Addr)
	if err := srv.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}
