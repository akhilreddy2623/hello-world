// Code generated by mockery v2.38.0. DO NOT EDIT.

package mocks

import (
	enums "geico.visualstudio.com/Billing/plutus/enums"
	db "geico.visualstudio.com/Billing/plutus/payment-administrator-common/models/db"

	messaging "geico.visualstudio.com/Billing/plutus/common-models/messaging"

	mock "github.com/stretchr/testify/mock"
)

// ProcessTaskHandlerInterface is an autogenerated mock type for the ProcessTaskHandlerInterface type
type ProcessTaskHandlerInterface struct {
	mock.Mock
}

// CallWorkdayAPI provides a mock function with given fields: workdayFeed
func (_m *ProcessTaskHandlerInterface) CallWorkdayAPI(workdayFeed []*db.WorkdayFeed) error {
	ret := _m.Called(workdayFeed)

	if len(ret) == 0 {
		panic("no return value specified for CallWorkdayAPI")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func([]*db.WorkdayFeed) error); ok {
		r0 = rf(workdayFeed)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// PublishTaskErrorResponse provides a mock function with given fields: executeTaskRequest, errorType, err, executetaskResponseTopic, count
func (_m *ProcessTaskHandlerInterface) PublishTaskErrorResponse(executeTaskRequest messaging.ExecuteTaskRequest, errorType string, err error, executetaskResponseTopic string, count int) error {
	ret := _m.Called(executeTaskRequest, errorType, err, executetaskResponseTopic, count)

	if len(ret) == 0 {
		panic("no return value specified for PublishTaskErrorResponse")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(messaging.ExecuteTaskRequest, string, error, string, int) error); ok {
		r0 = rf(executeTaskRequest, errorType, err, executetaskResponseTopic, count)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// PublishTaskResponse provides a mock function with given fields: executeTaskRequest, taskStatus, executeTaskResponseTopic, count
func (_m *ProcessTaskHandlerInterface) PublishTaskResponse(executeTaskRequest messaging.ExecuteTaskRequest, taskStatus enums.TaskStatus, executeTaskResponseTopic string, count int) error {
	ret := _m.Called(executeTaskRequest, taskStatus, executeTaskResponseTopic, count)

	if len(ret) == 0 {
		panic("no return value specified for PublishTaskResponse")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(messaging.ExecuteTaskRequest, enums.TaskStatus, string, int) error); ok {
		r0 = rf(executeTaskRequest, taskStatus, executeTaskResponseTopic, count)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// NewProcessTaskHandlerInterface creates a new instance of ProcessTaskHandlerInterface. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewProcessTaskHandlerInterface(t interface {
	mock.TestingT
	Cleanup(func())
}) *ProcessTaskHandlerInterface {
	mock := &ProcessTaskHandlerInterface{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
