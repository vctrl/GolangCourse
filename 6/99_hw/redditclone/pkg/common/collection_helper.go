package common

import (
	"context"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type CollectionHelper interface {
	Find(ctx context.Context, filter interface{},
		opts ...*options.FindOptions) (CursorHelper, error)
	FindOne(ctx context.Context, filter interface{},
		opts ...*options.FindOneOptions) SingleResultHelper
	FindOneAndUpdate(ctx context.Context, filter interface{}, update interface{},
		opts ...*options.FindOneAndUpdateOptions) SingleResultHelper
	InsertOne(ctx context.Context, document interface{},
		opts ...*options.InsertOneOptions) (InsertOneResultHelper, error)
	UpdateOne(ctx context.Context, filter interface{}, update interface{},
		opts ...*options.UpdateOptions) (UpdateResultHelper, error)
	DeleteOne(ctx context.Context, filter interface{},
		opts ...*options.DeleteOptions) (DeleteResultHelper, error)
}

type SingleResultHelper interface {
	Decode(v interface{}) error
}

type CursorHelper interface {
	Close(ctx context.Context) error
	All(ctx context.Context, results interface{}) error
}

type InsertOneResultHelper interface {
	GetInsertedID() interface{}
}

type UpdateResultHelper interface {
	GetModifiedCount() int64
}

type DeleteResultHelper interface {
	GetDeletedCount() int64
}

type MongoCollection struct {
	Collection *mongo.Collection
}

func (mc *MongoCollection) Find(ctx context.Context, filter interface{},
	opts ...*options.FindOptions) (CursorHelper, error) {
	cur, err := mc.Collection.Find(ctx, filter, opts...)
	return &MongoCursor{cur: cur}, err
}

type MongoCursor struct {
	cur *mongo.Cursor
}

func (mc *MongoCursor) Close(ctx context.Context) error {
	return mc.cur.Close(ctx)
}

func (mc *MongoCursor) All(ctx context.Context, results interface{}) error {
	return mc.cur.All(ctx, results)
}

func (mc *MongoCollection) FindOneAndUpdate(ctx context.Context, filter interface{}, update interface{},
	opts ...*options.FindOneAndUpdateOptions) SingleResultHelper {
	return &MongoSingleResult{sr: mc.Collection.FindOneAndUpdate(ctx, filter, update, opts...)}
}

type MongoSingleResult struct {
	sr *mongo.SingleResult
}

func (msr *MongoSingleResult) Decode(v interface{}) error {
	return msr.sr.Decode(v)
}

func (mc *MongoCollection) InsertOne(ctx context.Context, document interface{},
	opts ...*options.InsertOneOptions) (InsertOneResultHelper, error) {
	res, err := mc.Collection.InsertOne(ctx, document, opts...)
	return &MongoInsertOneResult{res: res}, err
}

type MongoInsertOneResult struct {
	res *mongo.InsertOneResult
}

func (r *MongoInsertOneResult) GetInsertedID() interface{} {
	return r.res.InsertedID.(primitive.ObjectID).Hex()
}

func (mc *MongoCollection) FindOne(ctx context.Context, filter interface{},
	opts ...*options.FindOneOptions) SingleResultHelper {
	res := mc.Collection.FindOne(ctx, filter, opts...)

	return &MongoSingleResult{sr: res}
}

func (mc *MongoCollection) UpdateOne(ctx context.Context, filter interface{}, update interface{},
	opts ...*options.UpdateOptions) (UpdateResultHelper, error) {
	res, err := mc.Collection.UpdateOne(ctx, filter, update, opts...)
	return &MongoUpdateResult{res: res}, err
}

func (mc *MongoCollection) DeleteOne(ctx context.Context, filter interface{},
	opts ...*options.DeleteOptions) (DeleteResultHelper, error) {
	res, err := mc.Collection.DeleteOne(ctx, filter, opts...)
	return &MongoDeleteResult{res: res}, err
}

type MongoDeleteResult struct {
	res *mongo.DeleteResult
}

func (r *MongoDeleteResult) GetDeletedCount() int64 {
	return r.res.DeletedCount
}

type MongoUpdateResult struct {
	res *mongo.UpdateResult
}

func (r *MongoUpdateResult) GetModifiedCount() int64 {
	return r.res.ModifiedCount
}
