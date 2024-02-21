// Code generated by MockGen. DO NOT EDIT.
// Source: destination_definition.go
//
// Generated by this command:
//
//	mockgen -source=destination_definition.go -destination=../../testhelper/mock/destination_definition.go -package=mock_test IDestination
//

// Package mock_test is a generated GoMock package.
package mock_test

import (
	reflect "reflect"

	option "github.com/couchbaselabs/cbmigrate/internal/couchbase/option"
	index "github.com/couchbaselabs/cbmigrate/internal/index"
	gomock "go.uber.org/mock/gomock"
)

// MockIDestination is a mock of IDestination interface.
type MockIDestination struct {
	ctrl     *gomock.Controller
	recorder *MockIDestinationMockRecorder
}

// MockIDestinationMockRecorder is the mock recorder for MockIDestination.
type MockIDestinationMockRecorder struct {
	mock *MockIDestination
}

// NewMockIDestination creates a new mock instance.
func NewMockIDestination(ctrl *gomock.Controller) *MockIDestination {
	mock := &MockIDestination{ctrl: ctrl}
	mock.recorder = &MockIDestinationMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockIDestination) EXPECT() *MockIDestinationMockRecorder {
	return m.recorder
}

// Complete mocks base method.
func (m *MockIDestination) Complete() error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Complete")
	ret0, _ := ret[0].(error)
	return ret0
}

// Complete indicates an expected call of Complete.
func (mr *MockIDestinationMockRecorder) Complete() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Complete", reflect.TypeOf((*MockIDestination)(nil).Complete))
}

// CreateIndexes mocks base method.
func (m *MockIDestination) CreateIndexes(indexes []index.Index) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CreateIndexes", indexes)
	ret0, _ := ret[0].(error)
	return ret0
}

// CreateIndexes indicates an expected call of CreateIndexes.
func (mr *MockIDestinationMockRecorder) CreateIndexes(indexes any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CreateIndexes", reflect.TypeOf((*MockIDestination)(nil).CreateIndexes), indexes)
}

// Init mocks base method.
func (m *MockIDestination) Init(opts *option.Options) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Init", opts)
	ret0, _ := ret[0].(error)
	return ret0
}

// Init indicates an expected call of Init.
func (mr *MockIDestinationMockRecorder) Init(opts any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Init", reflect.TypeOf((*MockIDestination)(nil).Init), opts)
}

// ProcessData mocks base method.
func (m *MockIDestination) ProcessData(arg0 map[string]any) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ProcessData", arg0)
	ret0, _ := ret[0].(error)
	return ret0
}

// ProcessData indicates an expected call of ProcessData.
func (mr *MockIDestinationMockRecorder) ProcessData(arg0 any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ProcessData", reflect.TypeOf((*MockIDestination)(nil).ProcessData), arg0)
}
