package mocks

import (
	gomock "github.com/golang/mock/gomock"
	reflect "reflect"
)

type MockNodeQuerier struct {
	ctrl     *gomock.Controller
	recorder *MockNodeQuerierMockRecorder
}

type MockNodeQuerierMockRecorder struct {
	mock *MockNodeQuerier
}

func NewMockNodeQuerier(ctrl *gomock.Controller) *MockNodeQuerier {
	mock := &MockNodeQuerier{ctrl: ctrl}
	mock.recorder = &MockNodeQuerierMockRecorder{mock}
	return mock
}

func (m *MockNodeQuerier) EXPECT() *MockNodeQuerierMockRecorder {
	return m.recorder
}

func (m *MockNodeQuerier) QueryWithData(path string, data []byte) ([]byte, int64, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "QueryWithData", path, data)
	ret0, _ := ret[0].([]byte)
	ret1, _ := ret[1].(int64)
	ret2, _ := ret[2].(error)
	return ret0, ret1, ret2
}

func (mr *MockNodeQuerierMockRecorder) QueryWithData(path, data interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "QueryWithData", reflect.TypeOf((*MockNodeQuerier)(nil).QueryWithData), path, data)
}
