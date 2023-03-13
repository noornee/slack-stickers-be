// Code generated by MockGen. DO NOT EDIT.
// Source: media.go

// Package media is a generated GoMock package.
package media

import (
	context "context"
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
	model "github.com/odetolakehinde/slack-stickers-be/src/model"
)

// MockService is a mock of Service interface.
type MockService struct {
	ctrl     *gomock.Controller
	recorder *MockServiceMockRecorder
}

// MockServiceMockRecorder is the mock recorder for MockService.
type MockServiceMockRecorder struct {
	mock *MockService
}

// NewMockService creates a new mock instance.
func NewMockService(ctrl *gomock.Controller) *MockService {
	mock := &MockService{ctrl: ctrl}
	mock.recorder = &MockServiceMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockService) EXPECT() *MockServiceMockRecorder {
	return m.recorder
}

// SearchByTag mocks base method.
func (m *MockService) SearchByTag(ctx context.Context, tag string) ([]*model.Sticker, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SearchByTag", ctx, tag)
	ret0, _ := ret[0].([]*model.Sticker)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// SearchByTag indicates an expected call of SearchByTag.
func (mr *MockServiceMockRecorder) SearchByTag(ctx, tag interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SearchByTag", reflect.TypeOf((*MockService)(nil).SearchByTag), ctx, tag)
}

// UploadSticker mocks base method.
func (m *MockService) UploadSticker(ctx context.Context, name, details string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "UploadSticker", ctx, name, details)
	ret0, _ := ret[0].(error)
	return ret0
}

// UploadSticker indicates an expected call of UploadSticker.
func (mr *MockServiceMockRecorder) UploadSticker(ctx, name, details interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UploadSticker", reflect.TypeOf((*MockService)(nil).UploadSticker), ctx, name, details)
}
