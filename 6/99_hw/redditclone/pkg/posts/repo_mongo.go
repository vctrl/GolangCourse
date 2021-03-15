package posts

import (
	"context"
	"fmt"
	"redditclone/pkg/common"
	"strconv"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type PostsRepoMongo struct {
	collection common.CollectionHelper
}

func NewMongoClient(ctx context.Context, uri string) (*mongo.Client, error) {
	return mongo.Connect(ctx, options.Client().ApplyURI(uri))
}

func NewPostsRepoMongo(client *mongo.Client) *PostsRepoMongo {
	return &PostsRepoMongo{collection: &common.MongoCollection{Collection: client.Database("redditclone_db").Collection("posts")}}
}

func (r *PostsRepoMongo) GetAll(ctx context.Context) ([]*Post, error) {
	return r.getByField(ctx, bson.M{})
}

func (r *PostsRepoMongo) GetByCategory(ctx context.Context, category string) ([]*Post, error) {
	return r.getByField(ctx, bson.M{"category": category})
}

func (r *PostsRepoMongo) GetByAuthorID(ctx context.Context, authorID interface{}) ([]*Post, error) {
	return r.getByField(ctx, bson.M{"authorID": authorID})
}

func (r *PostsRepoMongo) GetByID(ctx context.Context, id interface{}) (*Post, error) {
	res := r.collection.FindOneAndUpdate(ctx, bson.M{"_id": id},
		bson.D{
			{"$inc", bson.D{{"views", 1}}},
		})

	post := &Post{}
	err := res.Decode(post)
	if err != nil {
		return nil, err
	}

	post.Views++
	return post, nil
}

func (r *PostsRepoMongo) Add(ctx context.Context, p *Post) (interface{}, error) {
	p.Votes = map[int64]VoteValue{p.AuthorID: Upvote}
	res, err := r.collection.InsertOne(ctx, p)
	if err != nil {
		return 0, err
	}

	return res.GetInsertedID(), nil
}

func (r *PostsRepoMongo) Delete(ctx context.Context, id interface{}) (bool, error) {
	res, err := r.collection.DeleteOne(ctx, bson.M{"_id": id})
	if err != nil {
		return false, err
	}

	if res.GetDeletedCount() == 0 {
		return false, nil
	}

	return true, nil
}

func (r *PostsRepoMongo) Upvote(ctx context.Context, postID interface{}, userID int64) (*Post, error) {
	return r.vote(ctx, postID, userID, Upvote)
}

func (r *PostsRepoMongo) DownVote(ctx context.Context, postID interface{}, userID int64) (*Post, error) {
	return r.vote(ctx, postID, userID, Downvote)
}

func (r *PostsRepoMongo) Unvote(ctx context.Context, postID interface{}, userID int64) (*Post, error) {
	return r.vote(ctx, postID, userID, Unvote)
}

func (r *PostsRepoMongo) ParseID(in string) (interface{}, error) {
	return primitive.ObjectIDFromHex(in)
}

func (r *PostsRepoMongo) vote(ctx context.Context, postID interface{}, userID int64, v VoteValue) (*Post, error) {

	// if v == Unvote {
	// 	delete(p.Votes, userID)
	// } else {
	// 	p.Votes[userID] = v
	// }

	// updateRes, err := r.collection.UpdateOne(ctx, bson.M{"_id": postID},
	// 	bson.D{
	// 		{"$set", bson.D{{"votes", p.Votes}}},
	// 	})
	// if err != nil {
	// 	return nil, err
	// }

	var updateRes common.UpdateResultHelper
	var err error
	if v == Unvote {
		updateRes, err = r.collection.UpdateOne(ctx, bson.M{"_id": postID},
			bson.D{
				{"$unset", bson.D{{"votes", strconv.FormatInt(userID, 10)}}},
			})
	} else {
		updateRes, err = r.collection.UpdateOne(ctx, bson.M{"_id": postID},
			bson.D{
				{"$set", bson.D{{"votes", bson.M{strconv.FormatInt(userID, 10): v}}}},
			})
	}
	if err != nil {
		return nil, err
	}

	if updateRes.GetModifiedCount() == 0 {
		return nil, fmt.Errorf("can't update post")
	}

	res := r.collection.FindOne(ctx, bson.M{"_id": postID})
	p := &Post{}
	err = res.Decode(p)
	if err != nil {
		return nil, err
	}

	return p, nil
}

func (r *PostsRepoMongo) getByField(ctx context.Context, filter bson.M) ([]*Post, error) {
	cur, err := r.collection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}

	defer cur.Close(ctx)

	var posts []*Post
	err = cur.All(ctx, &posts)
	if err != nil {
		return nil, err
	}

	return posts, nil
}
