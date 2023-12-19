// Code generated by mockery. DO NOT EDIT.

package mockaction

import mock "github.com/stretchr/testify/mock"

// Console is an autogenerated mock type for the Console type
type Console struct {
	mock.Mock
}

type Console_Expecter struct {
	mock *mock.Mock
}

func (_m *Console) EXPECT() *Console_Expecter {
	return &Console_Expecter{mock: &_m.Mock}
}

// Confirm provides a mock function with given fields: s
func (_m *Console) Confirm(s string) bool {
	ret := _m.Called(s)

	var r0 bool
	if rf, ok := ret.Get(0).(func(string) bool); ok {
		r0 = rf(s)
	} else {
		r0 = ret.Get(0).(bool)
	}

	return r0
}

// Console_Confirm_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Confirm'
type Console_Confirm_Call struct {
	*mock.Call
}

// Confirm is a helper method to define mock.On call
//   - s string
func (_e *Console_Expecter) Confirm(s interface{}) *Console_Confirm_Call {
	return &Console_Confirm_Call{Call: _e.mock.On("Confirm", s)}
}

func (_c *Console_Confirm_Call) Run(run func(s string)) *Console_Confirm_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(string))
	})
	return _c
}

func (_c *Console_Confirm_Call) Return(_a0 bool) *Console_Confirm_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *Console_Confirm_Call) RunAndReturn(run func(string) bool) *Console_Confirm_Call {
	_c.Call.Return(run)
	return _c
}

// Error provides a mock function with given fields: message
func (_m *Console) Error(message string) {
	_m.Called(message)
}

// Console_Error_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Error'
type Console_Error_Call struct {
	*mock.Call
}

// Error is a helper method to define mock.On call
//   - message string
func (_e *Console_Expecter) Error(message interface{}) *Console_Error_Call {
	return &Console_Error_Call{Call: _e.mock.On("Error", message)}
}

func (_c *Console_Error_Call) Run(run func(message string)) *Console_Error_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(string))
	})
	return _c
}

func (_c *Console_Error_Call) Return() *Console_Error_Call {
	_c.Call.Return()
	return _c
}

func (_c *Console_Error_Call) RunAndReturn(run func(string)) *Console_Error_Call {
	_c.Call.Return(run)
	return _c
}

// ErrorLn provides a mock function with given fields: message
func (_m *Console) ErrorLn(message string) {
	_m.Called(message)
}

// Console_ErrorLn_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'ErrorLn'
type Console_ErrorLn_Call struct {
	*mock.Call
}

// ErrorLn is a helper method to define mock.On call
//   - message string
func (_e *Console_Expecter) ErrorLn(message interface{}) *Console_ErrorLn_Call {
	return &Console_ErrorLn_Call{Call: _e.mock.On("ErrorLn", message)}
}

func (_c *Console_ErrorLn_Call) Run(run func(message string)) *Console_ErrorLn_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(string))
	})
	return _c
}

func (_c *Console_ErrorLn_Call) Return() *Console_ErrorLn_Call {
	_c.Call.Return()
	return _c
}

func (_c *Console_ErrorLn_Call) RunAndReturn(run func(string)) *Console_ErrorLn_Call {
	_c.Call.Return(run)
	return _c
}

// Errorf provides a mock function with given fields: message, a
func (_m *Console) Errorf(message string, a ...interface{}) {
	var _ca []interface{}
	_ca = append(_ca, message)
	_ca = append(_ca, a...)
	_m.Called(_ca...)
}

// Console_Errorf_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Errorf'
type Console_Errorf_Call struct {
	*mock.Call
}

// Errorf is a helper method to define mock.On call
//   - message string
//   - a ...interface{}
func (_e *Console_Expecter) Errorf(message interface{}, a ...interface{}) *Console_Errorf_Call {
	return &Console_Errorf_Call{Call: _e.mock.On("Errorf",
		append([]interface{}{message}, a...)...)}
}

func (_c *Console_Errorf_Call) Run(run func(message string, a ...interface{})) *Console_Errorf_Call {
	_c.Call.Run(func(args mock.Arguments) {
		variadicArgs := make([]interface{}, len(args)-1)
		for i, a := range args[1:] {
			if a != nil {
				variadicArgs[i] = a.(interface{})
			}
		}
		run(args[0].(string), variadicArgs...)
	})
	return _c
}

func (_c *Console_Errorf_Call) Return() *Console_Errorf_Call {
	_c.Call.Return()
	return _c
}

func (_c *Console_Errorf_Call) RunAndReturn(run func(string, ...interface{})) *Console_Errorf_Call {
	_c.Call.Return(run)
	return _c
}

// Fatal provides a mock function with given fields: err
func (_m *Console) Fatal(err error) {
	_m.Called(err)
}

// Console_Fatal_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Fatal'
type Console_Fatal_Call struct {
	*mock.Call
}

// Fatal is a helper method to define mock.On call
//   - err error
func (_e *Console_Expecter) Fatal(err interface{}) *Console_Fatal_Call {
	return &Console_Fatal_Call{Call: _e.mock.On("Fatal", err)}
}

func (_c *Console_Fatal_Call) Run(run func(err error)) *Console_Fatal_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(error))
	})
	return _c
}

func (_c *Console_Fatal_Call) Return() *Console_Fatal_Call {
	_c.Call.Return()
	return _c
}

func (_c *Console_Fatal_Call) RunAndReturn(run func(error)) *Console_Fatal_Call {
	_c.Call.Return(run)
	return _c
}

// Info provides a mock function with given fields: message
func (_m *Console) Info(message string) {
	_m.Called(message)
}

// Console_Info_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Info'
type Console_Info_Call struct {
	*mock.Call
}

// Info is a helper method to define mock.On call
//   - message string
func (_e *Console_Expecter) Info(message interface{}) *Console_Info_Call {
	return &Console_Info_Call{Call: _e.mock.On("Info", message)}
}

func (_c *Console_Info_Call) Run(run func(message string)) *Console_Info_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(string))
	})
	return _c
}

func (_c *Console_Info_Call) Return() *Console_Info_Call {
	_c.Call.Return()
	return _c
}

func (_c *Console_Info_Call) RunAndReturn(run func(string)) *Console_Info_Call {
	_c.Call.Return(run)
	return _c
}

// InfoLn provides a mock function with given fields: message
func (_m *Console) InfoLn(message string) {
	_m.Called(message)
}

// Console_InfoLn_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'InfoLn'
type Console_InfoLn_Call struct {
	*mock.Call
}

// InfoLn is a helper method to define mock.On call
//   - message string
func (_e *Console_Expecter) InfoLn(message interface{}) *Console_InfoLn_Call {
	return &Console_InfoLn_Call{Call: _e.mock.On("InfoLn", message)}
}

func (_c *Console_InfoLn_Call) Run(run func(message string)) *Console_InfoLn_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(string))
	})
	return _c
}

func (_c *Console_InfoLn_Call) Return() *Console_InfoLn_Call {
	_c.Call.Return()
	return _c
}

func (_c *Console_InfoLn_Call) RunAndReturn(run func(string)) *Console_InfoLn_Call {
	_c.Call.Return(run)
	return _c
}

// Infof provides a mock function with given fields: message, a
func (_m *Console) Infof(message string, a ...interface{}) {
	var _ca []interface{}
	_ca = append(_ca, message)
	_ca = append(_ca, a...)
	_m.Called(_ca...)
}

// Console_Infof_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Infof'
type Console_Infof_Call struct {
	*mock.Call
}

// Infof is a helper method to define mock.On call
//   - message string
//   - a ...interface{}
func (_e *Console_Expecter) Infof(message interface{}, a ...interface{}) *Console_Infof_Call {
	return &Console_Infof_Call{Call: _e.mock.On("Infof",
		append([]interface{}{message}, a...)...)}
}

func (_c *Console_Infof_Call) Run(run func(message string, a ...interface{})) *Console_Infof_Call {
	_c.Call.Run(func(args mock.Arguments) {
		variadicArgs := make([]interface{}, len(args)-1)
		for i, a := range args[1:] {
			if a != nil {
				variadicArgs[i] = a.(interface{})
			}
		}
		run(args[0].(string), variadicArgs...)
	})
	return _c
}

func (_c *Console_Infof_Call) Return() *Console_Infof_Call {
	_c.Call.Return()
	return _c
}

func (_c *Console_Infof_Call) RunAndReturn(run func(string, ...interface{})) *Console_Infof_Call {
	_c.Call.Return(run)
	return _c
}

// NumberPlural provides a mock function with given fields: count, one, many
func (_m *Console) NumberPlural(count int, one string, many string) string {
	ret := _m.Called(count, one, many)

	var r0 string
	if rf, ok := ret.Get(0).(func(int, string, string) string); ok {
		r0 = rf(count, one, many)
	} else {
		r0 = ret.Get(0).(string)
	}

	return r0
}

// Console_NumberPlural_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'NumberPlural'
type Console_NumberPlural_Call struct {
	*mock.Call
}

// NumberPlural is a helper method to define mock.On call
//   - count int
//   - one string
//   - many string
func (_e *Console_Expecter) NumberPlural(count interface{}, one interface{}, many interface{}) *Console_NumberPlural_Call {
	return &Console_NumberPlural_Call{Call: _e.mock.On("NumberPlural", count, one, many)}
}

func (_c *Console_NumberPlural_Call) Run(run func(count int, one string, many string)) *Console_NumberPlural_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(int), args[1].(string), args[2].(string))
	})
	return _c
}

func (_c *Console_NumberPlural_Call) Return(_a0 string) *Console_NumberPlural_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *Console_NumberPlural_Call) RunAndReturn(run func(int, string, string) string) *Console_NumberPlural_Call {
	_c.Call.Return(run)
	return _c
}

// Success provides a mock function with given fields: message
func (_m *Console) Success(message string) {
	_m.Called(message)
}

// Console_Success_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Success'
type Console_Success_Call struct {
	*mock.Call
}

// Success is a helper method to define mock.On call
//   - message string
func (_e *Console_Expecter) Success(message interface{}) *Console_Success_Call {
	return &Console_Success_Call{Call: _e.mock.On("Success", message)}
}

func (_c *Console_Success_Call) Run(run func(message string)) *Console_Success_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(string))
	})
	return _c
}

func (_c *Console_Success_Call) Return() *Console_Success_Call {
	_c.Call.Return()
	return _c
}

func (_c *Console_Success_Call) RunAndReturn(run func(string)) *Console_Success_Call {
	_c.Call.Return(run)
	return _c
}

// SuccessLn provides a mock function with given fields: message
func (_m *Console) SuccessLn(message string) {
	_m.Called(message)
}

// Console_SuccessLn_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'SuccessLn'
type Console_SuccessLn_Call struct {
	*mock.Call
}

// SuccessLn is a helper method to define mock.On call
//   - message string
func (_e *Console_Expecter) SuccessLn(message interface{}) *Console_SuccessLn_Call {
	return &Console_SuccessLn_Call{Call: _e.mock.On("SuccessLn", message)}
}

func (_c *Console_SuccessLn_Call) Run(run func(message string)) *Console_SuccessLn_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(string))
	})
	return _c
}

func (_c *Console_SuccessLn_Call) Return() *Console_SuccessLn_Call {
	_c.Call.Return()
	return _c
}

func (_c *Console_SuccessLn_Call) RunAndReturn(run func(string)) *Console_SuccessLn_Call {
	_c.Call.Return(run)
	return _c
}

// Successf provides a mock function with given fields: message, a
func (_m *Console) Successf(message string, a ...interface{}) {
	var _ca []interface{}
	_ca = append(_ca, message)
	_ca = append(_ca, a...)
	_m.Called(_ca...)
}

// Console_Successf_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Successf'
type Console_Successf_Call struct {
	*mock.Call
}

// Successf is a helper method to define mock.On call
//   - message string
//   - a ...interface{}
func (_e *Console_Expecter) Successf(message interface{}, a ...interface{}) *Console_Successf_Call {
	return &Console_Successf_Call{Call: _e.mock.On("Successf",
		append([]interface{}{message}, a...)...)}
}

func (_c *Console_Successf_Call) Run(run func(message string, a ...interface{})) *Console_Successf_Call {
	_c.Call.Run(func(args mock.Arguments) {
		variadicArgs := make([]interface{}, len(args)-1)
		for i, a := range args[1:] {
			if a != nil {
				variadicArgs[i] = a.(interface{})
			}
		}
		run(args[0].(string), variadicArgs...)
	})
	return _c
}

func (_c *Console_Successf_Call) Return() *Console_Successf_Call {
	_c.Call.Return()
	return _c
}

func (_c *Console_Successf_Call) RunAndReturn(run func(string, ...interface{})) *Console_Successf_Call {
	_c.Call.Return(run)
	return _c
}

// Warn provides a mock function with given fields: message
func (_m *Console) Warn(message string) {
	_m.Called(message)
}

// Console_Warn_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Warn'
type Console_Warn_Call struct {
	*mock.Call
}

// Warn is a helper method to define mock.On call
//   - message string
func (_e *Console_Expecter) Warn(message interface{}) *Console_Warn_Call {
	return &Console_Warn_Call{Call: _e.mock.On("Warn", message)}
}

func (_c *Console_Warn_Call) Run(run func(message string)) *Console_Warn_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(string))
	})
	return _c
}

func (_c *Console_Warn_Call) Return() *Console_Warn_Call {
	_c.Call.Return()
	return _c
}

func (_c *Console_Warn_Call) RunAndReturn(run func(string)) *Console_Warn_Call {
	_c.Call.Return(run)
	return _c
}

// WarnLn provides a mock function with given fields: message
func (_m *Console) WarnLn(message string) {
	_m.Called(message)
}

// Console_WarnLn_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'WarnLn'
type Console_WarnLn_Call struct {
	*mock.Call
}

// WarnLn is a helper method to define mock.On call
//   - message string
func (_e *Console_Expecter) WarnLn(message interface{}) *Console_WarnLn_Call {
	return &Console_WarnLn_Call{Call: _e.mock.On("WarnLn", message)}
}

func (_c *Console_WarnLn_Call) Run(run func(message string)) *Console_WarnLn_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(string))
	})
	return _c
}

func (_c *Console_WarnLn_Call) Return() *Console_WarnLn_Call {
	_c.Call.Return()
	return _c
}

func (_c *Console_WarnLn_Call) RunAndReturn(run func(string)) *Console_WarnLn_Call {
	_c.Call.Return(run)
	return _c
}

// Warnf provides a mock function with given fields: message, a
func (_m *Console) Warnf(message string, a ...interface{}) {
	var _ca []interface{}
	_ca = append(_ca, message)
	_ca = append(_ca, a...)
	_m.Called(_ca...)
}

// Console_Warnf_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Warnf'
type Console_Warnf_Call struct {
	*mock.Call
}

// Warnf is a helper method to define mock.On call
//   - message string
//   - a ...interface{}
func (_e *Console_Expecter) Warnf(message interface{}, a ...interface{}) *Console_Warnf_Call {
	return &Console_Warnf_Call{Call: _e.mock.On("Warnf",
		append([]interface{}{message}, a...)...)}
}

func (_c *Console_Warnf_Call) Run(run func(message string, a ...interface{})) *Console_Warnf_Call {
	_c.Call.Run(func(args mock.Arguments) {
		variadicArgs := make([]interface{}, len(args)-1)
		for i, a := range args[1:] {
			if a != nil {
				variadicArgs[i] = a.(interface{})
			}
		}
		run(args[0].(string), variadicArgs...)
	})
	return _c
}

func (_c *Console_Warnf_Call) Return() *Console_Warnf_Call {
	_c.Call.Return()
	return _c
}

func (_c *Console_Warnf_Call) RunAndReturn(run func(string, ...interface{})) *Console_Warnf_Call {
	_c.Call.Return(run)
	return _c
}

type mockConstructorTestingTNewConsole interface {
	mock.TestingT
	Cleanup(func())
}

// NewConsole creates a new instance of Console. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
func NewConsole(t mockConstructorTestingTNewConsole) *Console {
	mock := &Console{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
