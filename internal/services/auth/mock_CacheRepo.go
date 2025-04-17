// Code generated by mockery. DO NOT EDIT.

package auth

import (
	context "context"
	time "time"

	mock "github.com/stretchr/testify/mock"
)

// MockCacheRepo is an autogenerated mock type for the CacheRepo type
type MockCacheRepo struct {
	mock.Mock
}

type MockCacheRepo_Expecter struct {
	mock *mock.Mock
}

func (_m *MockCacheRepo) EXPECT() *MockCacheRepo_Expecter {
	return &MockCacheRepo_Expecter{mock: &_m.Mock}
}

// Get provides a mock function with given fields: ctx, key
func (_m *MockCacheRepo) Get(ctx context.Context, key string) (string, error) {
	ret := _m.Called(ctx, key)

	if len(ret) == 0 {
		panic("no return value specified for Get")
	}

	var r0 string
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string) (string, error)); ok {
		return rf(ctx, key)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string) string); ok {
		r0 = rf(ctx, key)
	} else {
		r0 = ret.Get(0).(string)
	}

	if rf, ok := ret.Get(1).(func(context.Context, string) error); ok {
		r1 = rf(ctx, key)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// MockCacheRepo_Get_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Get'
type MockCacheRepo_Get_Call struct {
	*mock.Call
}

// Get is a helper method to define mock.On call
//   - ctx context.Context
//   - key string
func (_e *MockCacheRepo_Expecter) Get(ctx interface{}, key interface{}) *MockCacheRepo_Get_Call {
	return &MockCacheRepo_Get_Call{Call: _e.mock.On("Get", ctx, key)}
}

func (_c *MockCacheRepo_Get_Call) Run(run func(ctx context.Context, key string)) *MockCacheRepo_Get_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(string))
	})
	return _c
}

func (_c *MockCacheRepo_Get_Call) Return(_a0 string, _a1 error) *MockCacheRepo_Get_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *MockCacheRepo_Get_Call) RunAndReturn(run func(context.Context, string) (string, error)) *MockCacheRepo_Get_Call {
	_c.Call.Return(run)
	return _c
}

// Set provides a mock function with given fields: ctx, key, value, expiration
func (_m *MockCacheRepo) Set(ctx context.Context, key string, value any, expiration time.Duration) error {
	ret := _m.Called(ctx, key, value, expiration)

	if len(ret) == 0 {
		panic("no return value specified for Set")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, string, any, time.Duration) error); ok {
		r0 = rf(ctx, key, value, expiration)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// MockCacheRepo_Set_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Set'
type MockCacheRepo_Set_Call struct {
	*mock.Call
}

// Set is a helper method to define mock.On call
//   - ctx context.Context
//   - key string
//   - value any
//   - expiration time.Duration
func (_e *MockCacheRepo_Expecter) Set(ctx interface{}, key interface{}, value interface{}, expiration interface{}) *MockCacheRepo_Set_Call {
	return &MockCacheRepo_Set_Call{Call: _e.mock.On("Set", ctx, key, value, expiration)}
}

func (_c *MockCacheRepo_Set_Call) Run(run func(ctx context.Context, key string, value any, expiration time.Duration)) *MockCacheRepo_Set_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(string), args[2].(any), args[3].(time.Duration))
	})
	return _c
}

func (_c *MockCacheRepo_Set_Call) Return(_a0 error) *MockCacheRepo_Set_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *MockCacheRepo_Set_Call) RunAndReturn(run func(context.Context, string, any, time.Duration) error) *MockCacheRepo_Set_Call {
	_c.Call.Return(run)
	return _c
}

// NewMockCacheRepo creates a new instance of MockCacheRepo. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewMockCacheRepo(t interface {
	mock.TestingT
	Cleanup(func())
}) *MockCacheRepo {
	mock := &MockCacheRepo{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
