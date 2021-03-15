// Code generated by MockGen. DO NOT EDIT.
// Source: manager_jwt.go

// Package session is a generated GoMock package.
package session

import (
	context "context"
	gomock "github.com/golang/mock/gomock"
	http "net/http"
	reflect "reflect"
)

// MockSessionManager is a mock of SessionManager interface
type MockSessionManager struct {
	ctrl     *gomock.Controller
	recorder *MockSessionManagerMockRecorder
}

// MockSessionManagerMockRecorder is the mock recorder for MockSessionManager
type MockSessionManagerMockRecorder struct {
	mock *MockSessionManager
}

// NewMockSessionManager creates a new mock instance
func NewMockSessionManager(ctrl *gomock.Controller) *MockSessionManager {
	mock := &MockSessionManager{ctrl: ctrl}
	mock.recorder = &MockSessionManagerMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockSessionManager) EXPECT() *MockSessionManagerMockRecorder {
	return m.recorder
}

// Create mocks base method
func (m *MockSessionManager) Create(ctx context.Context, w http.ResponseWriter, u *User, sessID string, expiresAt int64) (string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Create", ctx, w, u, sessID, expiresAt)
	ret0, _ := ret[0].(string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Create indicates an expected call of Create
func (mr *MockSessionManagerMockRecorder) Create(ctx, w, u, sessID, expiresAt interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Create", reflect.TypeOf((*MockSessionManager)(nil).Create), ctx, w, u, sessID, expiresAt)
}

// Check mocks base method
func (m *MockSessionManager) Check(ctx context.Context, r *http.Request) (*Session, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Check", ctx, r)
	ret0, _ := ret[0].(*Session)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Check indicates an expected call of Check
func (mr *MockSessionManagerMockRecorder) Check(ctx, r interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Check", reflect.TypeOf((*MockSessionManager)(nil).Check), ctx, r)
}

// Destroy mocks base method
func (m *MockSessionManager) Destroy(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Destroy", ctx, w, r)
	ret0, _ := ret[0].(error)
	return ret0
}

// Destroy indicates an expected call of Destroy
func (mr *MockSessionManagerMockRecorder) Destroy(ctx, w, r interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Destroy", reflect.TypeOf((*MockSessionManager)(nil).Destroy), ctx, w, r)
}

// DestroyAll mocks base method
func (m *MockSessionManager) DestroyAll(ctx context.Context, user *User) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DestroyAll", ctx, user)
	ret0, _ := ret[0].(error)
	return ret0
}

// DestroyAll indicates an expected call of DestroyAll
func (mr *MockSessionManagerMockRecorder) DestroyAll(ctx, user interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DestroyAll", reflect.TypeOf((*MockSessionManager)(nil).DestroyAll), ctx, user)
}
