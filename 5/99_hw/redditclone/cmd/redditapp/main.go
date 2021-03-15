package main

import (
	"flag"
	"log"
	"net/http"
	"redditclone/pkg/comments"
	"redditclone/pkg/handlers"
	"redditclone/pkg/middleware"
	"redditclone/pkg/posts"
	"redditclone/pkg/session"
	"redditclone/pkg/user"
	"time"

	"github.com/gorilla/mux"
	"go.uber.org/zap"
)

func main() {
	var dir string

	flag.StringVar(&dir, "dir", "template", "the directory to serve files from. Defaults to the current dir")
	flag.Parse()
	r := mux.NewRouter()

	sm, err := session.NewSessionsJWTManager()
	if err != nil {
		log.Fatal(err.Error())
		return
	}

	zapLogger, _ := zap.NewProduction()
	defer zapLogger.Sync() // flushes buffer, if any
	logger := zapLogger.Sugar()

	userRepo := user.NewRepo()
	userHandler := &handlers.UserHandler{
		Sm:     sm,
		Repo:   userRepo,
		Logger: logger,
	}

	commentsRepo := comments.NewRepo()
	postsRepo := posts.NewRepo()
	postsHandler := &handlers.PostHandler{
		Sm:           sm,
		PostsRepo:    postsRepo,
		UsersRepo:    userRepo,
		Logger:       logger,
		CommentsRepo: commentsRepo,
	}

	commentsHandler := &handlers.CommentHandler{CommentsRepo: commentsRepo, PostsRepo: postsRepo, UsersRepo: userRepo, Logger: logger}
	api := r.PathPrefix("/api/").Subrouter()

	api.HandleFunc("/login", userHandler.Login).Methods(http.MethodGet)
	api.HandleFunc("/register", userHandler.Register).Methods(http.MethodPost)

	api.HandleFunc("/posts/", postsHandler.GetAll).Methods(http.MethodGet)
	api.HandleFunc("/posts", postsHandler.Create).Methods(http.MethodPost)
	api.HandleFunc("/posts/{category}", postsHandler.GetPostsByCategory).Methods(http.MethodGet)
	api.HandleFunc("/post/{id}", postsHandler.GetByID).Methods(http.MethodGet)
	api.HandleFunc("/post/{id}", postsHandler.Delete).Methods(http.MethodDelete)
	api.HandleFunc("/user/{username}", postsHandler.GetByUser).Methods(http.MethodGet)

	api.HandleFunc("/post/{post_id}/upvote", postsHandler.Upvote).Methods(http.MethodGet)
	api.HandleFunc("/post/{post_id}/downvote", postsHandler.Downvote).Methods(http.MethodGet)
	api.HandleFunc("/post/{post_id}/unvote", postsHandler.Unvote).Methods(http.MethodGet)

	api.HandleFunc("/post/{post_id}", commentsHandler.Add).Methods(http.MethodPost)
	api.HandleFunc("/post/{post_id}/{comment_id}", commentsHandler.Delete).Methods(http.MethodDelete)

	r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("./template/static"))))
	r.PathPrefix("/").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "template/index.html")
	})

	r.NotFoundHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handlers.WriteResponse(w, "not found", http.StatusNotFound)
	})

	mux := middleware.Auth(logger, sm, r)
	mux = middleware.Log(logger, mux)
	mux = middleware.Recover(logger, mux)

	srv := &http.Server{
		Handler:      mux,
		Addr:         "127.0.0.1:8000",
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	logger.Infof("Started server at %s", srv.Addr)
	log.Fatal(srv.ListenAndServe())
}
