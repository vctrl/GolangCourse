// Code generated by MockGen. DO NOT EDIT.
// Source: pkg/handlers/posts.go

// Package handlers is a generated GoMock package.
package handlers

import (
	context "context"
	gomock "github.com/golang/mock/gomock"
	posts "redditclone/pkg/posts"
	reflect "reflect"
)

// MockPostsRepo is a mock of PostsRepo interface
type MockPostsRepo struct {
	ctrl     *gomock.Controller
	recorder *MockPostsRepoMockRecorder
}

// MockPostsRepoMockRecorder is the mock recorder for MockPostsRepo
type MockPostsRepoMockRecorder struct {
	mock *MockPostsRepo
}

// NewMockPostsRepo creates a new mock instance
func NewMockPostsRepo(ctrl *gomock.Controller) *MockPostsRepo {
	mock := &MockPostsRepo{ctrl: ctrl}
	mock.recorder = &MockPostsRepoMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockPostsRepo) EXPECT() *MockPostsRepoMockRecorder {
	return m.recorder
}

// GetAll mocks base method
func (m *MockPostsRepo) GetAll(arg0 context.Context) ([]*posts.Post, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetAll", arg0)
	ret0, _ := ret[0].([]*posts.Post)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetAll indicates an expected call of GetAll
func (mr *MockPostsRepoMockRecorder) GetAll(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetAll", reflect.TypeOf((*MockPostsRepo)(nil).GetAll), arg0)
}

// GetByID mocks base method
func (m *MockPostsRepo) GetByID(arg0 context.Context, arg1 interface{}) (*posts.Post, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetByID", arg0, arg1)
	ret0, _ := ret[0].(*posts.Post)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetByID indicates an expected call of GetByID
func (mr *MockPostsRepoMockRecorder) GetByID(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetByID", reflect.TypeOf((*MockPostsRepo)(nil).GetByID), arg0, arg1)
}

// GetByCategory mocks base method
func (m *MockPostsRepo) GetByCategory(arg0 context.Context, arg1 string) ([]*posts.Post, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetByCategory", arg0, arg1)
	ret0, _ := ret[0].([]*posts.Post)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetByCategory indicates an expected call of GetByCategory
func (mr *MockPostsRepoMockRecorder) GetByCategory(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetByCategory", reflect.TypeOf((*MockPostsRepo)(nil).GetByCategory), arg0, arg1)
}

// GetByAuthorID mocks base method
func (m *MockPostsRepo) GetByAuthorID(arg0 context.Context, arg1 interface{}) ([]*posts.Post, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetByAuthorID", arg0, arg1)
	ret0, _ := ret[0].([]*posts.Post)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetByAuthorID indicates an expected call of GetByAuthorID
func (mr *MockPostsRepoMockRecorder) GetByAuthorID(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetByAuthorID", reflect.TypeOf((*MockPostsRepo)(nil).GetByAuthorID), arg0, arg1)
}

// Add mocks base method
func (m *MockPostsRepo) Add(arg0 context.Context, arg1 *posts.Post) (interface{}, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Add", arg0, arg1)
	ret0, _ := ret[0].(interface{})
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Add indicates an expected call of Add
func (mr *MockPostsRepoMockRecorder) Add(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Add", reflect.TypeOf((*MockPostsRepo)(nil).Add), arg0, arg1)
}

// Delete mocks base method
func (m *MockPostsRepo) Delete(arg0 context.Context, arg1 interface{}) (bool, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Delete", arg0, arg1)
	ret0, _ := ret[0].(bool)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Delete indicates an expected call of Delete
func (mr *MockPostsRepoMockRecorder) Delete(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Delete", reflect.TypeOf((*MockPostsRepo)(nil).Delete), arg0, arg1)
}

// Upvote mocks base method
func (m *MockPostsRepo) Upvote(arg0 context.Context, arg1 interface{}, arg2 int64) (*posts.Post, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Upvote", arg0, arg1, arg2)
	ret0, _ := ret[0].(*posts.Post)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Upvote indicates an expected call of Upvote
func (mr *MockPostsRepoMockRecorder) Upvote(arg0, arg1, arg2 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Upvote", reflect.TypeOf((*MockPostsRepo)(nil).Upvote), arg0, arg1, arg2)
}

// DownVote mocks base method
func (m *MockPostsRepo) DownVote(arg0 context.Context, arg1 interface{}, arg2 int64) (*posts.Post, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DownVote", arg0, arg1, arg2)
	ret0, _ := ret[0].(*posts.Post)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// DownVote indicates an expected call of DownVote
func (mr *MockPostsRepoMockRecorder) DownVote(arg0, arg1, arg2 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DownVote", reflect.TypeOf((*MockPostsRepo)(nil).DownVote), arg0, arg1, arg2)
}

// Unvote mocks base method
func (m *MockPostsRepo) Unvote(arg0 context.Context, arg1 interface{}, arg2 int64) (*posts.Post, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Unvote", arg0, arg1, arg2)
	ret0, _ := ret[0].(*posts.Post)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Unvote indicates an expected call of Unvote
func (mr *MockPostsRepoMockRecorder) Unvote(arg0, arg1, arg2 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Unvote", reflect.TypeOf((*MockPostsRepo)(nil).Unvote), arg0, arg1, arg2)
}

// ParseID mocks base method
func (m *MockPostsRepo) ParseID(arg0 string) (interface{}, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ParseID", arg0)
	ret0, _ := ret[0].(interface{})
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ParseID indicates an expected call of ParseID
func (mr *MockPostsRepoMockRecorder) ParseID(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ParseID", reflect.TypeOf((*MockPostsRepo)(nil).ParseID), arg0)
}
