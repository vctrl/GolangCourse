package main

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"reflect"
	"testing"
	"time"

	"github.com/go-redis/redis/v8"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

const (
	url = "http://127.0.0.1:8000"
)

type Case struct {
	Name               string
	Body               map[string]string
	Path               string
	Method             string
	Status             int
	Headers            map[string]string
	MatchExpected      func(*testing.T, map[string]interface{}) (string, bool)
	MatchExpectedError string
	NeedAuth           bool
	NeedPostID         bool
}

var token string
var postID string

type Post map[string]interface{}

var PostValue = Post{
	"author":           map[string]interface{}{"id": float64(1), "username": "test_user"},
	"category":         "music",
	"comments":         []interface{}{},
	"created":          "2021-02-16T21:45:02.420486+03:00",
	"id":               "602c12aec2591712925c3dc5",
	"score":            float64(1),
	"text":             "Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. Ut enim ad minim veniam, quis nostrud exercitation ullamco laboris nisi ut aliquip ex ea commodo consequat.",
	"title":            "Lorem ipsum dolor sit amet",
	"type":             "text",
	"upvotePercentage": float64(100),
	"views":            float64(0),
	"votes":            []interface{}{map[string]interface{}{"user": float64(1), "vote": float64(1)}},
}

// todo return errror
func (p1 *Post) Equals(p2 Post, incViews bool) (string, bool) {
	if incViews {
		views := (*p1)["views"].(float64) + 1
		(*p1)["views"] = views
	}
	postID = p2["id"].(string)

	// id and time created are not fixed values
	if p2["id"] == "" {
		return fmt.Sprintf("id should not be empty"), false
	}
	p2["id"] = (*p1)["id"]
	if _, e := time.Parse(time.RFC3339, p2["created"].(string)); e != nil {
		return fmt.Sprintf("invalid time value: %v", p2["created"]), false
	}

	p2["created"] = (*p1)["created"]
	if !reflect.DeepEqual(p2, *p1) {
		return fmt.Sprintf("expected %s but was %s ", *p1, p2), false
	}

	return "", true
}

var cases = []Case{
	{
		Name: "success register",
		Body: map[string]string{
			"username": "test_user",
			"password": "test_password",
		},
		Path:    "/api/register",
		Status:  http.StatusCreated,
		Method:  http.MethodPost,
		Headers: map[string]string{"Content-type": "application/json"},
		MatchExpected: func(t *testing.T, m map[string]interface{}) (string, bool) {
			tok, ok := m["token"]
			return "", ok && tok != ""
		},
	},
	{
		Name: "success login",
		Body: map[string]string{
			"username": "test_user",
			"password": "test_password",
		},
		Path:    "/api/login",
		Status:  http.StatusOK,
		Method:  http.MethodPost,
		Headers: map[string]string{"Content-type": "application/json"},
		MatchExpected: func(t *testing.T, m map[string]interface{}) (string, bool) {
			tok, ok := m["token"]
			if ok {
				token = tok.(string)
			}
			return "", ok && tok != ""
		},
	},
	{
		Name: "success post create",
		Body: map[string]string{
			"category": "music",
			"text":     "Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. Ut enim ad minim veniam, quis nostrud exercitation ullamco laboris nisi ut aliquip ex ea commodo consequat.",
			"title":    "Lorem ipsum dolor sit amet",
			"type":     "text",
		},
		Path:   "/api/posts",
		Status: http.StatusCreated,
		Method: http.MethodPost,
		Headers: map[string]string{
			"Content-type": "application/json",
		},
		MatchExpected: func(t *testing.T, m map[string]interface{}) (string, bool) {
			return PostValue.Equals(m, false)
		},
		NeedAuth: true,
	},
	{
		Name: "get post success",
		Body: map[string]string{
			"username": "test_user",
			"password": "test_password",
		},
		Path:    "/api/post/",
		Status:  http.StatusOK,
		Method:  http.MethodGet,
		Headers: map[string]string{"Content-type": "application/json"},
		MatchExpected: func(t *testing.T, m map[string]interface{}) (string, bool) {
			return PostValue.Equals(m, true)
		},
		NeedPostID: true,
	},

	{
		Name: "fail post create unauthorized",
		Body: map[string]string{
			"category": "music",
			"text":     "Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. Ut enim ad minim veniam, quis nostrud exercitation ullamco laboris nisi ut aliquip ex ea commodo consequat.",
			"title":    "Lorem ipsum dolor sit amet",
			"type":     "text",
		},
		Path:   "/api/posts",
		Status: http.StatusUnauthorized,
		Method: http.MethodPost,
		Headers: map[string]string{
			"Content-type": "application/json",
		},
		MatchExpected: func(t *testing.T, m map[string]interface{}) (string, bool) {
			expected := map[string]interface{}{
				"message": "unauthorized",
			}

			if !reflect.DeepEqual(m, expected) {
				return fmt.Sprintf("unexpected message body: %v, expected: %v", m, expected), false
			}

			return "", true
		},
		NeedAuth: false,
	},
}

func TestFunctionality(t *testing.T) {
	startServer()

	for i, tc := range cases {
		requestBody, err := json.Marshal(tc.Body)

		if err != nil {
			t.Fatalf("test case %d %s failed: %s", i, tc.Name, err.Error())
		}

		url1 := url + tc.Path
		if tc.NeedPostID {
			url1 += postID
		}

		request, err := http.NewRequest(tc.Method, url1, bytes.NewBuffer(requestBody))

		if err != nil {
			t.Fatalf("test case %d %s, error on preparing request: %s", i, tc.Name, err.Error())
		}

		for h, v := range tc.Headers {
			request.Header.Set(h, v)
		}

		if tc.NeedAuth {
			request.Header.Set("Authorization", token)
		}

		client := &http.Client{
			Timeout: time.Second * 10,
		}
		resp, err := client.Do(request)
		if err != nil {
			t.Fatalf("test case %d %s failed, error on request: %s", i, tc.Name, err.Error())
		}

		if resp.StatusCode != tc.Status {
			t.Fatalf("test case %d %s failed, unexpected status code: %d", i, tc.Name, resp.StatusCode)
		}

		respBytes, err := ioutil.ReadAll(resp.Body)

		fact := map[string]interface{}{}
		err = json.Unmarshal(respBytes, &fact)
		if err != nil {
			t.Fatalf("test case %d %s failed, invalid json: %s", i, tc.Name, err.Error())
		}

		message, eq := tc.MatchExpected(t, fact)
		if !eq {
			t.Fatalf("test case %d %s failed: %s", i, tc.Name, message)
		}
	}
}

func cleanupDBs(a *Application) {
	// sql cleanup
	db, err := sql.Open("mysql", a.MySQLConnectionString)
	if err != nil {
		panic(err.Error())
	}

	defer db.Close()
	err = db.Ping()
	if err != nil {
		panic(err)
	}

	cleanupQueries := []string{`DROP TABLE users;`,
		`CREATE TABLE users (
		id int(11) unsigned NOT NULL AUTO_INCREMENT,
		password  VARBINARY(100) NOT NULL,
		username VARCHAR(50) NOT NULL,
		PRIMARY KEY (id)
	) ENGINE=INNODB DEFAULT CHARSET=utf8;`}

	for _, q := range cleanupQueries {
		_, err = db.Exec(q)
	}
	if err != nil {
		panic(err)
	}

	// mongo cleanup
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	mongoClient, err := mongo.Connect(ctx, options.Client().ApplyURI(a.MongoConnectionString))

	if err != nil {
		panic(err)
	}
	err = mongoClient.Ping(ctx, readpref.Primary())
	if err != nil {
		panic(err)
	}

	mongoClient.Database("redditclone_db").Collection("posts").DeleteMany(ctx, bson.D{})
	mongoClient.Database("redditclone_db").Collection("comments").DeleteMany(ctx, bson.D{})

	// todo insertion
	mongoClient.Database("redditclone_db").Collection("posts").InsertOne(ctx, map[string]string{})
	comments := []interface{}{map[string]string{}}
	mongoClient.Database("redditclone_db").Collection("posts").InsertMany(ctx, comments)

	// redis cleanup
	rdb := redis.NewClient(a.RedisOptions)

	ctx, cancel = context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := rdb.Ping(ctx).Err(); err != nil {
		panic(err.Error())
	}

	err = rdb.FlushDB(ctx).Err()
	if err != nil {
		panic(err)
	}
}

func startServer() {
	// тут по-хорошему должны быть настройки с адресами тестовых БД
	a := &Application{
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
		ServerAddr: "127.0.0.1:8000",

		PrivateKeyLocation: "../../key.rsa",
		PublicKeyLocation:  "../../key.rsa.pub",
	}

	cleanupDBs(a)

	go func() {
		a.Run()
	}()

	client := &http.Client{
		Timeout: time.Second * 10,
	}

	for {
		time.Sleep(time.Second)

		log.Println("checking if server started...")
		resp, err := client.Get(url)
		if err != nil {
			log.Printf("failed: %v", err.Error())
			continue
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			log.Printf("request failed with status code %v\n", resp.StatusCode)
			continue
		}

		break
	}

	log.Println("server is started successfully")
}
