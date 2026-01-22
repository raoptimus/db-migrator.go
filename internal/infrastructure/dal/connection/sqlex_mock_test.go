/**
 * This file is part of the raoptimus/db-migrator.go library
 *
 * @copyright Copyright (c) Evgeniy Urvantsev
 * @license https://github.com/raoptimus/db-migrator.go/blob/master/LICENSE.md
 * @link https://github.com/raoptimus/db-migrator.go
 */

package connection

import (
	"context"

	"github.com/raoptimus/db-migrator.go/internal/infrastructure/sqlex"
	"github.com/stretchr/testify/mock"
)

// MockTx is a mock type for the sqlex.Tx type
type MockTx struct {
	mock.Mock
}

type MockTx_Expecter struct {
	mock *mock.Mock
}

func (_m *MockTx) EXPECT() *MockTx_Expecter {
	return &MockTx_Expecter{mock: &_m.Mock}
}

// Commit provides a mock function with no fields
func (_m *MockTx) Commit() error {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for Commit")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func() error); ok {
		r0 = rf()
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// MockTx_Commit_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Commit'
type MockTx_Commit_Call struct {
	*mock.Call
}

// Commit is a helper method to define mock.On call
func (_e *MockTx_Expecter) Commit() *MockTx_Commit_Call {
	return &MockTx_Commit_Call{Call: _e.mock.On("Commit")}
}

func (_c *MockTx_Commit_Call) Run(run func()) *MockTx_Commit_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *MockTx_Commit_Call) Return(_a0 error) *MockTx_Commit_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *MockTx_Commit_Call) RunAndReturn(run func() error) *MockTx_Commit_Call {
	_c.Call.Return(run)
	return _c
}

// ExecContext provides a mock function with given fields: ctx, query, args
func (_m *MockTx) ExecContext(ctx context.Context, query string, args ...any) (sqlex.Result, error) {
	var _ca []interface{}
	_ca = append(_ca, ctx, query)
	_ca = append(_ca, args...)
	ret := _m.Called(_ca...)

	if len(ret) == 0 {
		panic("no return value specified for ExecContext")
	}

	var r0 sqlex.Result
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string, ...any) (sqlex.Result, error)); ok {
		return rf(ctx, query, args...)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string, ...any) sqlex.Result); ok {
		r0 = rf(ctx, query, args...)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(sqlex.Result)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, string, ...any) error); ok {
		r1 = rf(ctx, query, args...)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// MockTx_ExecContext_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'ExecContext'
type MockTx_ExecContext_Call struct {
	*mock.Call
}

// ExecContext is a helper method to define mock.On call
//   - ctx context.Context
//   - query string
//   - args ...any
func (_e *MockTx_Expecter) ExecContext(ctx interface{}, query interface{}, args ...interface{}) *MockTx_ExecContext_Call {
	return &MockTx_ExecContext_Call{Call: _e.mock.On("ExecContext",
		append([]interface{}{ctx, query}, args...)...)}
}

func (_c *MockTx_ExecContext_Call) Run(run func(ctx context.Context, query string, args ...any)) *MockTx_ExecContext_Call {
	_c.Call.Run(func(args mock.Arguments) {
		variadicArgs := make([]any, len(args)-2)
		for i, a := range args[2:] {
			if a != nil {
				variadicArgs[i] = a.(any)
			}
		}
		run(args[0].(context.Context), args[1].(string), variadicArgs...)
	})
	return _c
}

func (_c *MockTx_ExecContext_Call) Return(_a0 sqlex.Result, _a1 error) *MockTx_ExecContext_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *MockTx_ExecContext_Call) RunAndReturn(run func(context.Context, string, ...any) (sqlex.Result, error)) *MockTx_ExecContext_Call {
	_c.Call.Return(run)
	return _c
}

// PrepareContext provides a mock function with given fields: ctx, query
func (_m *MockTx) PrepareContext(ctx context.Context, query string) (sqlex.Stmt, error) {
	ret := _m.Called(ctx, query)

	if len(ret) == 0 {
		panic("no return value specified for PrepareContext")
	}

	var r0 sqlex.Stmt
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string) (sqlex.Stmt, error)); ok {
		return rf(ctx, query)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string) sqlex.Stmt); ok {
		r0 = rf(ctx, query)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(sqlex.Stmt)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, string) error); ok {
		r1 = rf(ctx, query)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// MockTx_PrepareContext_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'PrepareContext'
type MockTx_PrepareContext_Call struct {
	*mock.Call
}

// PrepareContext is a helper method to define mock.On call
//   - ctx context.Context
//   - query string
func (_e *MockTx_Expecter) PrepareContext(ctx interface{}, query interface{}) *MockTx_PrepareContext_Call {
	return &MockTx_PrepareContext_Call{Call: _e.mock.On("PrepareContext", ctx, query)}
}

func (_c *MockTx_PrepareContext_Call) Run(run func(ctx context.Context, query string)) *MockTx_PrepareContext_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(string))
	})
	return _c
}

func (_c *MockTx_PrepareContext_Call) Return(_a0 sqlex.Stmt, _a1 error) *MockTx_PrepareContext_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *MockTx_PrepareContext_Call) RunAndReturn(run func(context.Context, string) (sqlex.Stmt, error)) *MockTx_PrepareContext_Call {
	_c.Call.Return(run)
	return _c
}

// Rollback provides a mock function with no fields
func (_m *MockTx) Rollback() error {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for Rollback")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func() error); ok {
		r0 = rf()
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// MockTx_Rollback_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Rollback'
type MockTx_Rollback_Call struct {
	*mock.Call
}

// Rollback is a helper method to define mock.On call
func (_e *MockTx_Expecter) Rollback() *MockTx_Rollback_Call {
	return &MockTx_Rollback_Call{Call: _e.mock.On("Rollback")}
}

func (_c *MockTx_Rollback_Call) Run(run func()) *MockTx_Rollback_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *MockTx_Rollback_Call) Return(_a0 error) *MockTx_Rollback_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *MockTx_Rollback_Call) RunAndReturn(run func() error) *MockTx_Rollback_Call {
	_c.Call.Return(run)
	return _c
}

// NewMockTx creates a new instance of MockTx. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewMockTx(t interface {
	mock.TestingT
	Cleanup(func())
}) *MockTx {
	m := &MockTx{}
	m.Mock.Test(t)

	t.Cleanup(func() { m.AssertExpectations(t) })

	return m
}

// MockRows is a mock type for the sqlex.Rows type
type MockRows struct {
	mock.Mock
}

type MockRows_Expecter struct {
	mock *mock.Mock
}

func (_m *MockRows) EXPECT() *MockRows_Expecter {
	return &MockRows_Expecter{mock: &_m.Mock}
}

// Close provides a mock function with no fields
func (_m *MockRows) Close() error {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for Close")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func() error); ok {
		r0 = rf()
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// MockRows_Close_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Close'
type MockRows_Close_Call struct {
	*mock.Call
}

// Close is a helper method to define mock.On call
func (_e *MockRows_Expecter) Close() *MockRows_Close_Call {
	return &MockRows_Close_Call{Call: _e.mock.On("Close")}
}

func (_c *MockRows_Close_Call) Run(run func()) *MockRows_Close_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *MockRows_Close_Call) Return(_a0 error) *MockRows_Close_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *MockRows_Close_Call) RunAndReturn(run func() error) *MockRows_Close_Call {
	_c.Call.Return(run)
	return _c
}

// Err provides a mock function with no fields
func (_m *MockRows) Err() error {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for Err")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func() error); ok {
		r0 = rf()
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// MockRows_Err_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Err'
type MockRows_Err_Call struct {
	*mock.Call
}

// Err is a helper method to define mock.On call
func (_e *MockRows_Expecter) Err() *MockRows_Err_Call {
	return &MockRows_Err_Call{Call: _e.mock.On("Err")}
}

func (_c *MockRows_Err_Call) Run(run func()) *MockRows_Err_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *MockRows_Err_Call) Return(_a0 error) *MockRows_Err_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *MockRows_Err_Call) RunAndReturn(run func() error) *MockRows_Err_Call {
	_c.Call.Return(run)
	return _c
}

// Next provides a mock function with no fields
func (_m *MockRows) Next() bool {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for Next")
	}

	var r0 bool
	if rf, ok := ret.Get(0).(func() bool); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(bool)
	}

	return r0
}

// MockRows_Next_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Next'
type MockRows_Next_Call struct {
	*mock.Call
}

// Next is a helper method to define mock.On call
func (_e *MockRows_Expecter) Next() *MockRows_Next_Call {
	return &MockRows_Next_Call{Call: _e.mock.On("Next")}
}

func (_c *MockRows_Next_Call) Run(run func()) *MockRows_Next_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *MockRows_Next_Call) Return(_a0 bool) *MockRows_Next_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *MockRows_Next_Call) RunAndReturn(run func() bool) *MockRows_Next_Call {
	_c.Call.Return(run)
	return _c
}

// Scan provides a mock function with given fields: dest
func (_m *MockRows) Scan(dest ...any) error {
	var _ca []interface{}
	_ca = append(_ca, dest...)
	ret := _m.Called(_ca...)

	if len(ret) == 0 {
		panic("no return value specified for Scan")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(...any) error); ok {
		r0 = rf(dest...)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// MockRows_Scan_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Scan'
type MockRows_Scan_Call struct {
	*mock.Call
}

// Scan is a helper method to define mock.On call
//   - dest ...any
func (_e *MockRows_Expecter) Scan(dest ...interface{}) *MockRows_Scan_Call {
	return &MockRows_Scan_Call{Call: _e.mock.On("Scan",
		append([]interface{}{}, dest...)...)}
}

func (_c *MockRows_Scan_Call) Run(run func(dest ...any)) *MockRows_Scan_Call {
	_c.Call.Run(func(args mock.Arguments) {
		variadicArgs := make([]any, len(args)-0)
		for i, a := range args[0:] {
			if a != nil {
				variadicArgs[i] = a.(any)
			}
		}
		run(variadicArgs...)
	})
	return _c
}

func (_c *MockRows_Scan_Call) Return(_a0 error) *MockRows_Scan_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *MockRows_Scan_Call) RunAndReturn(run func(...any) error) *MockRows_Scan_Call {
	_c.Call.Return(run)
	return _c
}

// NewMockRows creates a new instance of MockRows. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewMockRows(t interface {
	mock.TestingT
	Cleanup(func())
}) *MockRows {
	m := &MockRows{}
	m.Mock.Test(t)

	t.Cleanup(func() { m.AssertExpectations(t) })

	return m
}

// MockResult is a mock type for the sqlex.Result type
type MockResult struct {
	mock.Mock
}

type MockResult_Expecter struct {
	mock *mock.Mock
}

func (_m *MockResult) EXPECT() *MockResult_Expecter {
	return &MockResult_Expecter{mock: &_m.Mock}
}

// LastInsertId provides a mock function with no fields
func (_m *MockResult) LastInsertId() (int64, error) {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for LastInsertId")
	}

	var r0 int64
	var r1 error
	if rf, ok := ret.Get(0).(func() (int64, error)); ok {
		return rf()
	}
	if rf, ok := ret.Get(0).(func() int64); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(int64)
	}

	if rf, ok := ret.Get(1).(func() error); ok {
		r1 = rf()
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// MockResult_LastInsertId_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'LastInsertId'
type MockResult_LastInsertId_Call struct {
	*mock.Call
}

// LastInsertId is a helper method to define mock.On call
func (_e *MockResult_Expecter) LastInsertId() *MockResult_LastInsertId_Call {
	return &MockResult_LastInsertId_Call{Call: _e.mock.On("LastInsertId")}
}

func (_c *MockResult_LastInsertId_Call) Run(run func()) *MockResult_LastInsertId_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *MockResult_LastInsertId_Call) Return(_a0 int64, _a1 error) *MockResult_LastInsertId_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *MockResult_LastInsertId_Call) RunAndReturn(run func() (int64, error)) *MockResult_LastInsertId_Call {
	_c.Call.Return(run)
	return _c
}

// RowsAffected provides a mock function with no fields
func (_m *MockResult) RowsAffected() (int64, error) {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for RowsAffected")
	}

	var r0 int64
	var r1 error
	if rf, ok := ret.Get(0).(func() (int64, error)); ok {
		return rf()
	}
	if rf, ok := ret.Get(0).(func() int64); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(int64)
	}

	if rf, ok := ret.Get(1).(func() error); ok {
		r1 = rf()
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// MockResult_RowsAffected_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'RowsAffected'
type MockResult_RowsAffected_Call struct {
	*mock.Call
}

// RowsAffected is a helper method to define mock.On call
func (_e *MockResult_Expecter) RowsAffected() *MockResult_RowsAffected_Call {
	return &MockResult_RowsAffected_Call{Call: _e.mock.On("RowsAffected")}
}

func (_c *MockResult_RowsAffected_Call) Run(run func()) *MockResult_RowsAffected_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *MockResult_RowsAffected_Call) Return(_a0 int64, _a1 error) *MockResult_RowsAffected_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *MockResult_RowsAffected_Call) RunAndReturn(run func() (int64, error)) *MockResult_RowsAffected_Call {
	_c.Call.Return(run)
	return _c
}

// NewMockResult creates a new instance of MockResult. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewMockResult(t interface {
	mock.TestingT
	Cleanup(func())
}) *MockResult {
	m := &MockResult{}
	m.Mock.Test(t)

	t.Cleanup(func() { m.AssertExpectations(t) })

	return m
}
