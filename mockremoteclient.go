// Automatically generated by MockGen. DO NOT EDIT!
// Source: remoteclient.go

package proxy

import (
	context "context"
	gomock "github.com/golang/mock/gomock"
)

// Mock of RemoteClient interface
type MockRemoteClient struct {
	ctrl     *gomock.Controller
	recorder *_MockRemoteClientRecorder
}

// Recorder for MockRemoteClient (not exported)
type _MockRemoteClientRecorder struct {
	mock *MockRemoteClient
}

func NewMockRemoteClient(ctrl *gomock.Controller) *MockRemoteClient {
	mock := &MockRemoteClient{ctrl: ctrl}
	mock.recorder = &_MockRemoteClientRecorder{mock}
	return mock
}

func (_m *MockRemoteClient) EXPECT() *_MockRemoteClientRecorder {
	return _m.recorder
}

func (_m *MockRemoteClient) Update(ctx context.Context, addr string, info string) error {
	ret := _m.ctrl.Call(_m, "Update", ctx, addr, info)
	ret0, _ := ret[0].(error)
	return ret0
}

func (_mr *_MockRemoteClientRecorder) Update(arg0, arg1, arg2 interface{}) *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "Update", arg0, arg1, arg2)
}