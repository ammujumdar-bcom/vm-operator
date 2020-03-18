// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/vmware-tanzu/vm-operator/pkg/vmprovider/providers/vsphere (interfaces: ContentDownloadHandler)

// Package mocks is a generated GoMock package.
package mocks

import (
	context "context"
	gomock "github.com/golang/mock/gomock"
	library "github.com/vmware/govmomi/vapi/library"
	rest "github.com/vmware/govmomi/vapi/rest"
	vsphere "github.com/vmware-tanzu/vm-operator/pkg/vmprovider/providers/vsphere"
	reflect "reflect"
)

// MockContentDownloadHandler is a mock of ContentDownloadHandler interface
type MockContentDownloadHandler struct {
	ctrl     *gomock.Controller
	recorder *MockContentDownloadHandlerMockRecorder
}

// MockContentDownloadHandlerMockRecorder is the mock recorder for MockContentDownloadHandler
type MockContentDownloadHandlerMockRecorder struct {
	mock *MockContentDownloadHandler
}

// NewMockContentDownloadHandler creates a new mock instance
func NewMockContentDownloadHandler(ctrl *gomock.Controller) *MockContentDownloadHandler {
	mock := &MockContentDownloadHandler{ctrl: ctrl}
	mock.recorder = &MockContentDownloadHandlerMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockContentDownloadHandler) EXPECT() *MockContentDownloadHandlerMockRecorder {
	return m.recorder
}

// GenerateDownloadUriForLibraryItem mocks base method
func (m *MockContentDownloadHandler) GenerateDownloadUriForLibraryItem(arg0 context.Context, arg1 *rest.Client, arg2 *library.Item) (vsphere.DownloadUriResponse, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GenerateDownloadUriForLibraryItem", arg0, arg1, arg2)
	ret0, _ := ret[0].(vsphere.DownloadUriResponse)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GenerateDownloadUriForLibraryItem indicates an expected call of GenerateDownloadUriForLibraryItem
func (mr *MockContentDownloadHandlerMockRecorder) GenerateDownloadUriForLibraryItem(arg0, arg1, arg2 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GenerateDownloadUriForLibraryItem", reflect.TypeOf((*MockContentDownloadHandler)(nil).GenerateDownloadUriForLibraryItem), arg0, arg1, arg2)
}
