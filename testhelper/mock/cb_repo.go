// Code generated by MockGen. DO NOT EDIT.
// Source: repo.go
//
// Generated by this command:
//
//	mockgen -source=repo.go -destination=../../../testhelper/mock/cb_repo.go -package=mock_test -mock_names=IRepo=MockCouchbaseIRepo IRepo
//

// Package mock_test is a generated GoMock package.
package mock_test

import (
	reflect "reflect"

	gocb "github.com/couchbase/gocb/v2"
	common "github.com/couchbaselabs/cbmigrate/internal/common"
	option "github.com/couchbaselabs/cbmigrate/internal/couchbase/option"
	gomock "go.uber.org/mock/gomock"
)

// MockCouchbaseIRepo is a mock of IRepo interface.
type MockCouchbaseIRepo struct {
	ctrl     *gomock.Controller
	recorder *MockCouchbaseIRepoMockRecorder
}

// MockCouchbaseIRepoMockRecorder is the mock recorder for MockCouchbaseIRepo.
type MockCouchbaseIRepoMockRecorder struct {
	mock *MockCouchbaseIRepo
}

// NewMockCouchbaseIRepo creates a new mock instance.
func NewMockCouchbaseIRepo(ctrl *gomock.Controller) *MockCouchbaseIRepo {
	mock := &MockCouchbaseIRepo{ctrl: ctrl}
	mock.recorder = &MockCouchbaseIRepoMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockCouchbaseIRepo) EXPECT() *MockCouchbaseIRepoMockRecorder {
	return m.recorder
}

// CreateCollection mocks base method.
func (m *MockCouchbaseIRepo) CreateCollection(scope, name string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CreateCollection", scope, name)
	ret0, _ := ret[0].(error)
	return ret0
}

// CreateCollection indicates an expected call of CreateCollection.
func (mr *MockCouchbaseIRepoMockRecorder) CreateCollection(scope, name any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CreateCollection", reflect.TypeOf((*MockCouchbaseIRepo)(nil).CreateCollection), scope, name)
}

// CreateIndex mocks base method.
func (m *MockCouchbaseIRepo) CreateIndex(scope, collection string, index common.Index) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CreateIndex", scope, collection, index)
	ret0, _ := ret[0].(error)
	return ret0
}

// CreateIndex indicates an expected call of CreateIndex.
func (mr *MockCouchbaseIRepoMockRecorder) CreateIndex(scope, collection, index any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CreateIndex", reflect.TypeOf((*MockCouchbaseIRepo)(nil).CreateIndex), scope, collection, index)
}

// CreateScope mocks base method.
func (m *MockCouchbaseIRepo) CreateScope(name string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CreateScope", name)
	ret0, _ := ret[0].(error)
	return ret0
}

// CreateScope indicates an expected call of CreateScope.
func (mr *MockCouchbaseIRepoMockRecorder) CreateScope(name any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CreateScope", reflect.TypeOf((*MockCouchbaseIRepo)(nil).CreateScope), name)
}

// GetAllScopes mocks base method.
func (m *MockCouchbaseIRepo) GetAllScopes() ([]gocb.ScopeSpec, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetAllScopes")
	ret0, _ := ret[0].([]gocb.ScopeSpec)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetAllScopes indicates an expected call of GetAllScopes.
func (mr *MockCouchbaseIRepoMockRecorder) GetAllScopes() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetAllScopes", reflect.TypeOf((*MockCouchbaseIRepo)(nil).GetAllScopes))
}

// Init mocks base method.
func (m *MockCouchbaseIRepo) Init(uri string, opts *option.Options) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Init", uri, opts)
	ret0, _ := ret[0].(error)
	return ret0
}

// Init indicates an expected call of Init.
func (mr *MockCouchbaseIRepoMockRecorder) Init(uri, opts any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Init", reflect.TypeOf((*MockCouchbaseIRepo)(nil).Init), uri, opts)
}

// UpsertData mocks base method.
func (m *MockCouchbaseIRepo) UpsertData(scope, collection string, docs []gocb.BulkOp) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "UpsertData", scope, collection, docs)
	ret0, _ := ret[0].(error)
	return ret0
}

// UpsertData indicates an expected call of UpsertData.
func (mr *MockCouchbaseIRepoMockRecorder) UpsertData(scope, collection, docs any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UpsertData", reflect.TypeOf((*MockCouchbaseIRepo)(nil).UpsertData), scope, collection, docs)
}
