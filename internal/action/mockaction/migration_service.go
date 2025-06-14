// Code generated by mockery. DO NOT EDIT.

package mockaction

import (
	context "context"

	entity "github.com/raoptimus/db-migrator.go/internal/dal/entity"
	mock "github.com/stretchr/testify/mock"
)

// MigrationService is an autogenerated mock type for the MigrationService type
type MigrationService struct {
	mock.Mock
}

type MigrationService_Expecter struct {
	mock *mock.Mock
}

func (_m *MigrationService) EXPECT() *MigrationService_Expecter {
	return &MigrationService_Expecter{mock: &_m.Mock}
}

// ApplyFile provides a mock function with given fields: ctx, _a1, fileName, safely
func (_m *MigrationService) ApplyFile(ctx context.Context, _a1 *entity.Migration, fileName string, safely bool) error {
	ret := _m.Called(ctx, _a1, fileName, safely)

	if len(ret) == 0 {
		panic("no return value specified for ApplyFile")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, *entity.Migration, string, bool) error); ok {
		r0 = rf(ctx, _a1, fileName, safely)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// MigrationService_ApplyFile_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'ApplyFile'
type MigrationService_ApplyFile_Call struct {
	*mock.Call
}

// ApplyFile is a helper method to define mock.On call
//   - ctx context.Context
//   - _a1 *entity.Migration
//   - fileName string
//   - safely bool
func (_e *MigrationService_Expecter) ApplyFile(ctx interface{}, _a1 interface{}, fileName interface{}, safely interface{}) *MigrationService_ApplyFile_Call {
	return &MigrationService_ApplyFile_Call{Call: _e.mock.On("ApplyFile", ctx, _a1, fileName, safely)}
}

func (_c *MigrationService_ApplyFile_Call) Run(run func(ctx context.Context, _a1 *entity.Migration, fileName string, safely bool)) *MigrationService_ApplyFile_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(*entity.Migration), args[2].(string), args[3].(bool))
	})
	return _c
}

func (_c *MigrationService_ApplyFile_Call) Return(_a0 error) *MigrationService_ApplyFile_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *MigrationService_ApplyFile_Call) RunAndReturn(run func(context.Context, *entity.Migration, string, bool) error) *MigrationService_ApplyFile_Call {
	_c.Call.Return(run)
	return _c
}

// Migrations provides a mock function with given fields: ctx, limit
func (_m *MigrationService) Migrations(ctx context.Context, limit int) (entity.Migrations, error) {
	ret := _m.Called(ctx, limit)

	if len(ret) == 0 {
		panic("no return value specified for Migrations")
	}

	var r0 entity.Migrations
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, int) (entity.Migrations, error)); ok {
		return rf(ctx, limit)
	}
	if rf, ok := ret.Get(0).(func(context.Context, int) entity.Migrations); ok {
		r0 = rf(ctx, limit)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(entity.Migrations)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, int) error); ok {
		r1 = rf(ctx, limit)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// MigrationService_Migrations_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Migrations'
type MigrationService_Migrations_Call struct {
	*mock.Call
}

// Migrations is a helper method to define mock.On call
//   - ctx context.Context
//   - limit int
func (_e *MigrationService_Expecter) Migrations(ctx interface{}, limit interface{}) *MigrationService_Migrations_Call {
	return &MigrationService_Migrations_Call{Call: _e.mock.On("Migrations", ctx, limit)}
}

func (_c *MigrationService_Migrations_Call) Run(run func(ctx context.Context, limit int)) *MigrationService_Migrations_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(int))
	})
	return _c
}

func (_c *MigrationService_Migrations_Call) Return(_a0 entity.Migrations, _a1 error) *MigrationService_Migrations_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *MigrationService_Migrations_Call) RunAndReturn(run func(context.Context, int) (entity.Migrations, error)) *MigrationService_Migrations_Call {
	_c.Call.Return(run)
	return _c
}

// NewMigrations provides a mock function with given fields: ctx
func (_m *MigrationService) NewMigrations(ctx context.Context) (entity.Migrations, error) {
	ret := _m.Called(ctx)

	if len(ret) == 0 {
		panic("no return value specified for NewMigrations")
	}

	var r0 entity.Migrations
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context) (entity.Migrations, error)); ok {
		return rf(ctx)
	}
	if rf, ok := ret.Get(0).(func(context.Context) entity.Migrations); ok {
		r0 = rf(ctx)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(entity.Migrations)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context) error); ok {
		r1 = rf(ctx)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// MigrationService_NewMigrations_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'NewMigrations'
type MigrationService_NewMigrations_Call struct {
	*mock.Call
}

// NewMigrations is a helper method to define mock.On call
//   - ctx context.Context
func (_e *MigrationService_Expecter) NewMigrations(ctx interface{}) *MigrationService_NewMigrations_Call {
	return &MigrationService_NewMigrations_Call{Call: _e.mock.On("NewMigrations", ctx)}
}

func (_c *MigrationService_NewMigrations_Call) Run(run func(ctx context.Context)) *MigrationService_NewMigrations_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context))
	})
	return _c
}

func (_c *MigrationService_NewMigrations_Call) Return(_a0 entity.Migrations, _a1 error) *MigrationService_NewMigrations_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *MigrationService_NewMigrations_Call) RunAndReturn(run func(context.Context) (entity.Migrations, error)) *MigrationService_NewMigrations_Call {
	_c.Call.Return(run)
	return _c
}

// RevertFile provides a mock function with given fields: ctx, _a1, fileName, safely
func (_m *MigrationService) RevertFile(ctx context.Context, _a1 *entity.Migration, fileName string, safely bool) error {
	ret := _m.Called(ctx, _a1, fileName, safely)

	if len(ret) == 0 {
		panic("no return value specified for RevertFile")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, *entity.Migration, string, bool) error); ok {
		r0 = rf(ctx, _a1, fileName, safely)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// MigrationService_RevertFile_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'RevertFile'
type MigrationService_RevertFile_Call struct {
	*mock.Call
}

// RevertFile is a helper method to define mock.On call
//   - ctx context.Context
//   - _a1 *entity.Migration
//   - fileName string
//   - safely bool
func (_e *MigrationService_Expecter) RevertFile(ctx interface{}, _a1 interface{}, fileName interface{}, safely interface{}) *MigrationService_RevertFile_Call {
	return &MigrationService_RevertFile_Call{Call: _e.mock.On("RevertFile", ctx, _a1, fileName, safely)}
}

func (_c *MigrationService_RevertFile_Call) Run(run func(ctx context.Context, _a1 *entity.Migration, fileName string, safely bool)) *MigrationService_RevertFile_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(*entity.Migration), args[2].(string), args[3].(bool))
	})
	return _c
}

func (_c *MigrationService_RevertFile_Call) Return(_a0 error) *MigrationService_RevertFile_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *MigrationService_RevertFile_Call) RunAndReturn(run func(context.Context, *entity.Migration, string, bool) error) *MigrationService_RevertFile_Call {
	_c.Call.Return(run)
	return _c
}

// NewMigrationService creates a new instance of MigrationService. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewMigrationService(t interface {
	mock.TestingT
	Cleanup(func())
}) *MigrationService {
	mock := &MigrationService{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
