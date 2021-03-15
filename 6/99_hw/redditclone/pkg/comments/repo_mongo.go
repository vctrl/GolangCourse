package comments

import (
	"context"
	"redditclone/pkg/common"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type CommentRepoMongo struct {
	collection common.CollectionHelper
}

func NewCommentsRepoMongo(db *mongo.Database) *CommentRepoMongo {
	return &CommentRepoMongo{collection: &common.MongoCollection{Collection: db.Collection("comments")}}
}

func (repo *CommentRepoMongo) GetByPostID(ctx context.Context, id interface{}) ([]*Comment, error) {
	cur, err := repo.collection.Find(ctx, bson.M{"postID": id})
	defer cur.Close(ctx)

	if err != nil {
		return nil, err
	}

	var comments []*Comment
	err = cur.All(ctx, &comments)
	if err != nil {
		return nil, err
	}

	return comments, nil
}

func (repo *CommentRepoMongo) GetByID(ctx context.Context, id interface{}) (*Comment, error) {
	// not implemented
	return nil, nil
}

func (repo *CommentRepoMongo) Add(ctx context.Context, comment *Comment) (interface{}, error) {
	res, err := repo.collection.InsertOne(ctx, comment)
	if err != nil {
		return nil, err
	}

	return res.GetInsertedID(), nil
}

func (repo *CommentRepoMongo) Delete(ctx context.Context, id interface{}) (bool, error) {
	res, err := repo.collection.DeleteOne(ctx, bson.M{"_id": id})
	if err != nil {
		return false, err
	}

	if res.GetDeletedCount() == 0 {
		return false, nil
	}

	return true, nil
}

func (repo *CommentRepoMongo) ParseID(in string) (interface{}, error) {
	return primitive.ObjectIDFromHex(in)
}
