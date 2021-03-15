package posts

import (
	"context"
	"errors"
	"redditclone/pkg/common"
	"reflect"
	"strconv"
	"testing"
	"time"

	gomock "github.com/golang/mock/gomock"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type getByFieldCase struct {
	name      string
	cond      bson.M
	cursorErr error
	findErr   error
	f         func(ctx context.Context, r *PostsRepoMongo) ([]*Post, error)
}

const (
	cat = "test_category"
)

var authorID = primitive.NewObjectID()

var getByFieldCases = []getByFieldCase{
	{
		name: "GetAllHappyCase",
		cond: bson.M{},
		f: func(ctx context.Context, r *PostsRepoMongo) ([]*Post, error) {
			return r.GetAll(ctx)
		},
	},
	{
		name: "GetByCategoryHappyCase",
		cond: bson.M{"category": cat},
		f: func(ctx context.Context, r *PostsRepoMongo) ([]*Post, error) {
			return r.GetByCategory(ctx, cat)
		},
	},
	{
		name: "GetByAuthorIDHappyCase",
		cond: bson.M{"authorID": authorID},
		f: func(ctx context.Context, r *PostsRepoMongo) ([]*Post, error) {
			return r.GetByAuthorID(ctx, authorID)
		},
	},
	{
		name:    "FindErrorExpected",
		cond:    bson.M{},
		findErr: errors.New("error while calling find"),
		f: func(ctx context.Context, r *PostsRepoMongo) ([]*Post, error) {
			return r.GetAll(ctx)
		},
	},
	{
		name:      "CursorErrorExpected",
		cond:      bson.M{},
		cursorErr: errors.New("cursor error"),
		f: func(ctx context.Context, r *PostsRepoMongo) ([]*Post, error) {
			return r.GetAll(ctx)
		},
	},
}

func TestGetByField(t *testing.T) {
	for i, c := range getByFieldCases {
		ctrl := gomock.NewController(t)
		mockCollection := common.NewMockCollectionHelper(ctrl)
		mockCursor := common.NewMockCursorHelper(ctrl)
		repo := &PostsRepoMongo{collection: mockCollection}

		ctx := context.Background()

		expectedPosts := []*Post{
			{ID: primitive.NewObjectID(), Score: 1, Views: 123, Type: Text, Title: "test title 1", AuthorID: int64(1), Category: Fashion, Text: "test", Created: time.Now(), Votes: map[int64]VoteValue{}},
			{ID: primitive.NewObjectID(), Score: 50, Views: 124, Type: Text, Title: "test title 2", AuthorID: int64(1), Category: Fashion, Text: "test", Created: time.Now(), Votes: map[int64]VoteValue{}},
			{ID: primitive.NewObjectID(), Score: 100, Views: 432, Type: Link, Title: "test title 3", AuthorID: int64(1), Category: Fashion, URL: "https://mail.ru/", Created: time.Now(), Votes: map[int64]VoteValue{}},
		}

		expectedFilter := c.cond

		mockCollection.EXPECT().Find(ctx, gomock.Eq(expectedFilter)).Return(mockCursor, c.cursorErr)
		mockCursor.EXPECT().All(ctx, gomock.AssignableToTypeOf(&expectedPosts)).
			SetArg(1, expectedPosts).Return(c.findErr)
		mockCursor.EXPECT().Close(ctx).Return(nil)

		res, err := c.f(ctx, repo)

		// todo refactor?
		if c.cursorErr != nil {
			if c.cursorErr != err {
				t.Errorf("test #%d %s fail, expected error: %v, but was %v", i, c.name, c.cursorErr, err)
			}
		} else if c.findErr != nil {
			if c.findErr != err {
				t.Errorf("test #%d %s fail, expected error: %v, but was %v", i, c.name, c.findErr, err)
			}
		} else if !reflect.DeepEqual(res, expectedPosts) {
			t.Errorf("test #%d %s fail, expected: %v, but was: %v", i, c.name, expectedPosts, res)
		}
	}
}

type voteCase struct {
	name         string
	votes        map[int64]VoteValue
	vote         func(repo *PostsRepoMongo, ctx context.Context, postID interface{}, authorID int64) (*Post, error)
	expected     map[int64]VoteValue
	findOneErr   error
	updateOneErr error
}

var userID = int64(1)
var voteCases = []voteCase{
	{
		name:  "UpvoteHappyCase",
		votes: map[int64]VoteValue{},
		vote: func(repo *PostsRepoMongo, ctx context.Context, postID interface{}, authorID int64) (*Post, error) {
			return repo.Upvote(ctx, postID, authorID)
		},
		expected: map[int64]VoteValue{userID: Upvote},
	},
	{
		name:  "DownvoteHappyCase",
		votes: map[int64]VoteValue{},
		vote: func(repo *PostsRepoMongo, ctx context.Context, postID interface{}, authorID int64) (*Post, error) {
			return repo.DownVote(ctx, postID, authorID)
		},
		expected: map[int64]VoteValue{userID: Downvote},
	},
	{
		name:  "UnvoteHappyCase",
		votes: map[int64]VoteValue{userID: Upvote},
		vote: func(repo *PostsRepoMongo, ctx context.Context, postID interface{}, authorID int64) (*Post, error) {
			return repo.Unvote(ctx, postID, authorID)
		},
		expected: map[int64]VoteValue{},
	},
	{
		name:  "FindOneErrorExpected",
		votes: map[int64]VoteValue{},
		vote: func(repo *PostsRepoMongo, ctx context.Context, postID interface{}, authorID int64) (*Post, error) {
			return repo.Upvote(ctx, postID, authorID)
		},
		expected:   map[int64]VoteValue{userID: Upvote},
		findOneErr: errors.New("error while calling collection.findOne"),
	},
	{
		name:  "UpdateOneErrorExpected",
		votes: map[int64]VoteValue{},
		vote: func(repo *PostsRepoMongo, ctx context.Context, postID interface{}, authorID int64) (*Post, error) {
			return repo.Upvote(ctx, postID, authorID)
		},
		expected:     map[int64]VoteValue{userID: Upvote},
		updateOneErr: errors.New("error while calling collection.updateOne"),
	},
}

func TestVotes(t *testing.T) {
	for i, c := range voteCases {
		ctrl := gomock.NewController(t)
		mockCollection := common.NewMockCollectionHelper(ctrl)

		mockFindOneResult := common.NewMockSingleResultHelper(ctrl)
		mockUpdateResult := common.NewMockUpdateResultHelper(ctrl)

		testRepo := &PostsRepoMongo{collection: mockCollection}
		ctx := context.Background()

		postID := primitive.NewObjectID()

		bsonM := bson.M{"_id": postID}

		unsetBsonD := bson.D{
			{"$unset", bson.D{{"votes", strconv.FormatInt(userID, 10)}}},
		}
		m := bson.M{strconv.FormatInt(userID, 10): c.expected[userID]}
		setBsonD := bson.D{
			{"$set", bson.D{{"votes", m}}},
		}

		if c.name == "UnvoteHappyCase" {
			mockCollection.EXPECT().UpdateOne(ctx, gomock.Eq(bsonM), gomock.Eq(unsetBsonD)).
				Return(mockUpdateResult, c.updateOneErr)
		} else {
			mockCollection.EXPECT().UpdateOne(ctx, gomock.Eq(bsonM), gomock.Eq(setBsonD)).
				Return(mockUpdateResult, c.updateOneErr)
		}

		expectedPost := &Post{ID: postID, Votes: c.expected}
		mockCollection.EXPECT().FindOne(ctx, gomock.Eq(bsonM)).
			Return(mockFindOneResult)
		mockFindOneResult.EXPECT().Decode(gomock.AssignableToTypeOf(expectedPost)).
			SetArg(0, *expectedPost).Return(c.findOneErr)

		mockUpdateResult.EXPECT().GetModifiedCount().Return(int64(1))

		res, err := c.vote(testRepo, ctx, postID, userID)

		if c.findOneErr != nil {
			if err != c.findOneErr {
				t.Errorf("test #%d %s fail, expected error: %v, but was %v", i, c.name, c.findOneErr, err)
			}
		} else if c.updateOneErr != nil {
			if err != c.updateOneErr {
				t.Errorf("test #%d %s fail, expected error: %v, but was %v", i, c.name, c.findOneErr, err)
			}
		} else {
			if !reflect.DeepEqual(res.Votes, c.expected) {
				t.Errorf("test #%d fail, expected: %v, but was: %v", i, c.expected, expectedPost.Votes)
			}
		}
	}
}

func TestGetByID(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockCollection := common.NewMockCollectionHelper(ctrl)
	mockSingleCollection := common.NewMockSingleResultHelper(ctrl)

	repo := &PostsRepoMongo{collection: mockCollection}
	ctx := context.Background()

	id := primitive.NewObjectID()
	bsonM := bson.M{"_id": id}
	bsonD := bson.D{{"$inc", bson.D{{"views", 1}}}}
	mockCollection.EXPECT().
		FindOneAndUpdate(ctx, gomock.AssignableToTypeOf(bsonM), gomock.AssignableToTypeOf(bsonD)).
		Return(mockSingleCollection)

	expectedPost := &Post{ID: id, Score: 1, Views: 123, Type: Text, Title: "test title 1", AuthorID: int64(1), Category: Fashion, Text: "test", Created: time.Now(), Votes: map[int64]VoteValue{}}
	mockSingleCollection.EXPECT().Decode(gomock.AssignableToTypeOf(expectedPost)).
		SetArg(0, *expectedPost).Return(nil)

	// todo check err
	res, _ := repo.GetByID(ctx, id)

	expectedPost.Views++
	if !reflect.DeepEqual(res, expectedPost) {
		t.Errorf("test fail, expected: %v, but was: %v", res, expectedPost)
	}
}

func TestAdd(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockCollection := common.NewMockCollectionHelper(ctrl)
	mockInsertOneResult := common.NewMockInsertOneResultHelper(ctrl)

	repo := &PostsRepoMongo{collection: mockCollection}
	ctx := context.Background()

	expectedInsertID := 1
	authorID := int64(256)
	expectedPost := &Post{AuthorID: authorID}
	mockCollection.EXPECT().InsertOne(ctx, gomock.AssignableToTypeOf(expectedPost)).
		Return(mockInsertOneResult, nil)

	mockInsertOneResult.EXPECT().GetInsertedID().Return(expectedInsertID)

	// todo check err
	res, _ := repo.Add(ctx, expectedPost)

	if !reflect.DeepEqual(res, expectedInsertID) {
		t.Errorf("test fail, expected: %v, but was: %v", res, expectedInsertID)
	}

	if !reflect.DeepEqual(expectedPost.Votes, map[int64]VoteValue{authorID: Upvote}) {
		t.Errorf("test fail, added post should be upvoted by author")
	}
}

func TestDelete(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockCollection := common.NewMockCollectionHelper(ctrl)
	mockDeleteResult := common.NewMockDeleteResultHelper(ctrl)

	repo := &PostsRepoMongo{collection: mockCollection}

	ctx := context.Background()

	id := primitive.NewObjectID()
	bsonM := bson.M{"_id": id}
	mockCollection.EXPECT().DeleteOne(ctx, gomock.AssignableToTypeOf(bsonM)).
		Return(mockDeleteResult, nil)

	mockDeleteResult.EXPECT().GetDeletedCount().Return(int64(1))

	// todo check err
	deleted, _ := repo.Delete(ctx, id)

	if !deleted {
		t.Error("test fail, expected true but was false")
	}
}

func TestParseID(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockCollection := common.NewMockCollectionHelper(ctrl)
	testRepo := &PostsRepoMongo{collection: mockCollection}

	id := primitive.NewObjectID()
	parsed, err := testRepo.ParseID(id.Hex())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	objID, ok := parsed.(primitive.ObjectID)
	if !ok {
		t.Fatalf("unexpected type: %t", parsed)
	}

	if objID.Hex() != id.Hex() {
		t.Fatalf("values not equal: %v, %v", objID.Hex(), id.Hex())
	}
}
