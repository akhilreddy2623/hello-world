// Code generated by mockery v2.38.0. DO NOT EDIT.

package mocks

import (
	messaging "geico.visualstudio.com/Billing/plutus/common-models/messaging"
	mock "github.com/stretchr/testify/mock"
)

// ExecutePaymentResponseHandlerInterface is an autogenerated mock type for the ExecutePaymentResponseHandlerInterface type
type ExecutePaymentResponseHandlerInterface struct {
	mock.Mock
}

// PublishPaymentResponse provides a mock function with given fields: executePaymentResponse
func (_m *ExecutePaymentResponseHandlerInterface) PublishPaymentResponse(executePaymentResponse messaging.ExecutePaymentResponse) error {
	ret := _m.Called(executePaymentResponse)

	if len(ret) == 0 {
		panic("no return value specified for PublishPaymentResponse")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(messaging.ExecutePaymentResponse) error); ok {
		r0 = rf(executePaymentResponse)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// NewExecutePaymentResponseHandlerInterface creates a new instance of ExecutePaymentResponseHandlerInterface. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewExecutePaymentResponseHandlerInterface(t interface {
	mock.TestingT
	Cleanup(func())
}) *ExecutePaymentResponseHandlerInterface {
	mock := &ExecutePaymentResponseHandlerInterface{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
