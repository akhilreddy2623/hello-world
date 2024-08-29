// Code generated by mockery v2.42.0. DO NOT EDIT.

package mocks

import (
	db "geico.visualstudio.com/Billing/plutus/config-manager-common/models/db"
	mock "github.com/stretchr/testify/mock"
)

// ConfigRepositoryInterface is an autogenerated mock type for the ConfigRepositoryInterface type
type ConfigRepositoryInterface struct {
	mock.Mock
}

// GetConfig provides a mock function with given fields: applicationName, environment
func (_m *ConfigRepositoryInterface) GetConfig(applicationName string, environment string) (*[]db.ConfigResponse, error) {
	ret := _m.Called(applicationName, environment)

	if len(ret) == 0 {
		panic("no return value specified for GetConfig")
	}

	var r0 *[]db.ConfigResponse
	var r1 error
	if rf, ok := ret.Get(0).(func(string, string) (*[]db.ConfigResponse, error)); ok {
		return rf(applicationName, environment)
	}
	if rf, ok := ret.Get(0).(func(string, string) *[]db.ConfigResponse); ok {
		r0 = rf(applicationName, environment)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*[]db.ConfigResponse)
		}
	}

	if rf, ok := ret.Get(1).(func(string, string) error); ok {
		r1 = rf(applicationName, environment)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// NewConfigRepositoryInterface creates a new instance of ConfigRepositoryInterface. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewConfigRepositoryInterface(t interface {
	mock.TestingT
	Cleanup(func())
}) *ConfigRepositoryInterface {
	mock := &ConfigRepositoryInterface{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
