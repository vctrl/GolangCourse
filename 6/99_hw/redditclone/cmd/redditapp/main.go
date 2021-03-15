package main

import (
	"context"
	"database/sql"
	"flag"
	"io/ioutil"
	"log"
	"net/http"
	"redditclone/pkg/comments"
	"redditclone/pkg/handlers"
	"redditclone/pkg/middleware"
	"redditclone/pkg/posts"
	"redditclone/pkg/session"
	"redditclone/pkg/user"
	"time"

	"github.com/go-redis/redis/v8"
	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"go.uber.org/zap"
)

const (
	createSchema = `CREATE TABLE IF NOT EXISTS users (
		id int(11) unsigned NOT NULL AUTO_INCREMENT,
		password  VARBINARY(100) NOT NULL,
		username VARCHAR(50) NOT NULL,
		PRIMARY KEY (id)
	) ENGINE=INNODB DEFAULT CHARSET=utf8;`
)

func main() {
	app := &Application{
		MongoConnectionString:   "mongodb://admin:password@localhost:2712/redditclone_db?authSource=redditclone_db&readPreference=primary&gssapiServiceName=mongodb&appname=redditclone&ssl=false",
		MongoDBName:             "redditclone_db",
		MongoPostsCollection:    "posts",
		MongoCommentsCollection: "comments",
		MySQLConnectionString:   "root:qwer1234@tcp(localhost:3306)/redditclone",
		RedisOptions: &redis.Options{
			Addr:     "localhost:6379",
			Password: "redis",
			DB:       0,
		},
		ServerAddr:         "127.0.0.1:8000",
		PrivateKeyLocation: "key.rsa",
		PublicKeyLocation:  "key.rsa.pub",
	}

	app.Run()
}

type Application struct {
	MongoConnectionString   string
	MongoDBName             string
	MongoPostsCollection    string
	MongoCommentsCollection string
	MySQLConnectionString   string
	RedisOptions            *redis.Options

	ServerAddr         string
	PublicKeyLocation  string
	PrivateKeyLocation string

	HTTPServer *http.Server
}

func (a *Application) Run() {
	var dir string

	flag.StringVar(&dir, "dir", "template", "the directory to serve files from. Defaults to the current dir")
	flag.Parse()
	r := mux.NewRouter()

	rdb := redis.NewClient(a.RedisOptions)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	if err := rdb.Ping(ctx).Err(); err != nil {
		panic(err.Error())
	}

	privateKeyBytes, err := ioutil.ReadFile(a.PrivateKeyLocation)
	if err != nil {
		panic(err)
	}

	publicKeyBytes, err := ioutil.ReadFile(a.PublicKeyLocation)
	if err != nil {
		panic(err)
	}

	smJWT, err := session.NewSessionsJWTManager(privateKeyBytes, publicKeyBytes)
	if err != nil {
		panic(err)
	}

	sm := session.NewSessionManagerRedis(rdb, smJWT)
	zapLogger, _ := zap.NewProduction()
	defer zapLogger.Sync() // flushes buffer, if any
	logger := zapLogger.Sugar()

	db, err := sql.Open("mysql", a.MySQLConnectionString)
	if err != nil {
		panic(err.Error())
	}

	defer db.Close()
	err = db.Ping()
	if err != nil {
		panic(err)
	}

	_, err = db.Exec(createSchema)
	if err != nil {
		panic(err)
	}

	userRepo := user.NewUserRepoSQL(db)

	userHandler := &handlers.UserHandler{
		Sm:     sm,
		Repo:   userRepo,
		Logger: logger,
	}

	ctx, cancel = context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	client, err := posts.NewMongoClient(ctx, a.MongoConnectionString)
	err = client.Ping(ctx, readpref.Primary())
	if err != nil {
		panic(err)
	}

	ctx, cancel = context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	err = client.Ping(ctx, nil)
	if err != nil {
		panic(err)
	}

	commentsRepo := comments.NewCommentsRepoMongo(client.Database(a.MongoDBName))

	postsRepo := posts.NewPostsRepoMongo(client)

	postsHandler := &handlers.PostHandler{
		Sm:           sm,
		PostsRepo:    postsRepo,
		UsersRepo:    userRepo,
		Logger:       logger,
		CommentsRepo: commentsRepo,
	}

	commentsHandler := &handlers.CommentHandler{CommentsRepo: commentsRepo, PostsRepo: postsRepo, UsersRepo: userRepo, Logger: logger}
	api := r.PathPrefix("/api/").Subrouter()

	api.HandleFunc("/login", userHandler.Login).Methods(http.MethodPost)
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

	api.NotFoundHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handlers.WriteResponse(w, "not found", http.StatusNotFound)
	})

	r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("./template/static"))))
	r.PathPrefix("/").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "template/index.html")
	})

	mux := middleware.Auth(logger, sm, r)
	mux = middleware.Log(logger, mux)
	mux = middleware.Recover(logger, mux)

	srv := &http.Server{
		Handler:      mux,
		Addr:         a.ServerAddr,
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}
	a.HTTPServer = srv

	logger.Infof("Started server at %s", srv.Addr)
	log.Fatal(srv.ListenAndServe())
}
