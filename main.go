package main

import (
	"log"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/rgynn/klottr/pkg/api"
	"github.com/rgynn/klottr/pkg/config"
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

	e := echo.New()

	e.Debug = cfg.Debug
	e.Server.IdleTimeout = cfg.IdleTimeout
	e.Server.ReadTimeout = cfg.ReadTimeout
	e.Server.WriteTimeout = cfg.WriteTimeout

	e.Use(
		middleware.RequestID(),
		middleware.Logger(),
		middleware.BodyDump(api.BodyDumpFunc),
		middleware.Gzip(),
		middleware.BodyLimit("100K"),
		middleware.CORSWithConfig(middleware.CORSConfig{
			AllowOrigins: cfg.CORSAllowOrigins,
			AllowHeaders: []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept},
		}),
	)

	e.POST("/auth/signin", api.SignInHandler)
	e.POST("/auth/signup", api.SignUpHandler)
	e.DELETE("/auth/deactivate", api.DeactivateHandler)

	v1 := e.Group("/api/1.0")

	v1.Use(
		middleware.JWTWithConfig(middleware.JWTConfig{
			SigningKey:  []byte(cfg.JWTSecret),
			TokenLookup: "header:" + echo.HeaderAuthorization,
			AuthScheme:  "Bearer",
		}),
	)

	// Users
	v1.GET("/users", api.SearchUsersHandler)
	v1.GET("/users/:username", api.GetAdminUserHandler)

	// Admin Users
	v1.POST("/admin/users", api.CreateAdminUserHandler)
	v1.GET("/admin/users", api.SearchAdminUsersHandler)
	v1.GET("/admin/users/:username", api.GetAdminUserHandler)
	v1.DELETE("/admin/users/:username", api.DeleteAdminUserHandler)

	// Threads
	v1.POST("/:category", api.CreateCategoryThreadHandler)
	v1.GET("/:category", api.ListCategoryThreadsHandler)
	v1.GET("/:category/:thread_id", api.GetCategoryThreadHandler)
	v1.POST("/:category/:thread_id/upvote", api.UpVoteCategoryThreadHandler)
	v1.POST("/:category/:thread_id/downvote", api.DownVoteCategoryThreadHandler)

	// Comments
	v1.POST("/:category/:thread_id/comments", api.CreateCommentHandler)
	v1.GET("/:category/:thread_id/comments/:comment_id", api.GetCommentHandler)
	v1.DELETE("/:category/:thread_id/comments/:comment_id", api.DeleteCommentHandler)
	v1.POST("/:category/:thread_id/comments/:comment_id/upvote", api.UpVoteCommentHandler)
	v1.POST("/:category/:thread_id/comments/:comment_id/downvote", api.DownVoteCommentHandler)

	if err := e.Start(cfg.Addr); err != nil {
		log.Fatal(err)
	}
}
