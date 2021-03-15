package comments

import (
	"context"
	"redditclone/pkg/common"
	"reflect"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Case struct {
	name           string
	cond           bson.M
	cursorErr      error
	findErr        error
	f              func(ctx context.Context, r *CommentRepoMongo) (interface{}, error)
	expectedResult interface{}
}

var id = primitive.NewObjectID()

var expectedComments = []*Comment{
	{
		Created:  time.Now(),
		AuthorID: int64(1),
		Body:     "some comment about something",
		ID:       id,
		PostID:   primitive.NewObjectID(),
	},
}

var cases = []Case{
	{
		name: "GetByPostIDHappyCase",
		cond: bson.M{"postID": id},
		f: func(ctx context.Context, r *CommentRepoMongo) (interface{}, error) {
			return r.GetByPostID(ctx, id)
		},
		expectedResult: expectedComments,
	},
	{
		name: "InsertCommentHappyCase",
		cond: bson.M{"postID": id},
		f: func(ctx context.Context, r *CommentRepoMongo) (interface{}, error) {
			return r.Add(ctx, expectedComments[0])
		},
		expectedResult: expectedComments[0].ID,
	},
	{
		name: "DeleteCommentHappyCase",
		cond: bson.M{"postID": id},
		f: func(ctx context.Context, r *CommentRepoMongo) (interface{}, error) {
			return r.Delete(ctx, id)
		},
		expectedResult: true,
	},
}

func TestGetByPostID(t *testing.T) {
	for i, tc := range cases {
		ctrl := gomock.NewController(t)
		mockCollection := common.NewMockCollectionHelper(ctrl)

		mockCursor := common.NewMockCursorHelper(ctrl)
		repo := &CommentRepoMongo{collection: mockCollection}

		ctx := context.Background()

		expectedFilter := tc.cond

		// GetByPostID

		mockCollection.EXPECT().Find(ctx, gomock.Eq(expectedFilter)).Return(mockCursor, nil)
		mockCursor.EXPECT().All(ctx, gomock.AssignableToTypeOf(&expectedComments)).
			SetArg(1, expectedComments).Return(nil)
		mockCursor.EXPECT().Close(ctx).Return(nil)

		// Add
		mockInsertOneResult := common.NewMockInsertOneResultHelper(ctrl)

		mockCollection.EXPECT().
			InsertOne(ctx, gomock.AssignableToTypeOf(expectedComments[0])).
			Return(mockInsertOneResult, nil)
		mockInsertOneResult.EXPECT().GetInsertedID().Return(expectedComments[0].ID)

		// Delete
		mockDeleteResult := common.NewMockDeleteResultHelper(ctrl)
		mockCollection.EXPECT().
			DeleteOne(ctx, gomock.Any()).
			Return(mockDeleteResult, nil)
		mockDeleteResult.EXPECT().GetDeletedCount().Return(int64(1))

		res, _ := tc.f(ctx, repo)

		if !reflect.DeepEqual(res, tc.expectedResult) {
			t.Errorf("test case %d %s failed, expected %s but was %s", i, tc.name, tc.expectedResult, res)
		}
	}
}
