// Code generated by mockery v2.46.3. DO NOT EDIT.

package mocks

import mock "github.com/stretchr/testify/mock"

// DisplayManager is an autogenerated mock type for the DisplayManager type
type DisplayManager struct {
	mock.Mock
}

type DisplayManager_Expecter struct {
	mock *mock.Mock
}

func (_m *DisplayManager) EXPECT() *DisplayManager_Expecter {
	return &DisplayManager_Expecter{mock: &_m.Mock}
}

// ShowCommand provides a mock function with given fields: command
func (_m *DisplayManager) ShowCommand(command string) {
	_m.Called(command)
}

// DisplayManager_ShowCommand_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'ShowCommand'
type DisplayManager_ShowCommand_Call struct {
	*mock.Call
}

// ShowCommand is a helper method to define mock.On call
//   - command string
func (_e *DisplayManager_Expecter) ShowCommand(command interface{}) *DisplayManager_ShowCommand_Call {
	return &DisplayManager_ShowCommand_Call{Call: _e.mock.On("ShowCommand", command)}
}

func (_c *DisplayManager_ShowCommand_Call) Run(run func(command string)) *DisplayManager_ShowCommand_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(string))
	})
	return _c
}

func (_c *DisplayManager_ShowCommand_Call) Return() *DisplayManager_ShowCommand_Call {
	_c.Call.Return()
	return _c
}

func (_c *DisplayManager_ShowCommand_Call) RunAndReturn(run func(string)) *DisplayManager_ShowCommand_Call {
	_c.Call.Return(run)
	return _c
}

// ShowError provides a mock function with given fields: message
func (_m *DisplayManager) ShowError(message string) {
	_m.Called(message)
}

// DisplayManager_ShowError_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'ShowError'
type DisplayManager_ShowError_Call struct {
	*mock.Call
}

// ShowError is a helper method to define mock.On call
//   - message string
func (_e *DisplayManager_Expecter) ShowError(message interface{}) *DisplayManager_ShowError_Call {
	return &DisplayManager_ShowError_Call{Call: _e.mock.On("ShowError", message)}
}

func (_c *DisplayManager_ShowError_Call) Run(run func(message string)) *DisplayManager_ShowError_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(string))
	})
	return _c
}

func (_c *DisplayManager_ShowError_Call) Return() *DisplayManager_ShowError_Call {
	_c.Call.Return()
	return _c
}

func (_c *DisplayManager_ShowError_Call) RunAndReturn(run func(string)) *DisplayManager_ShowError_Call {
	_c.Call.Return(run)
	return _c
}

// ShowInfo provides a mock function with given fields: message
func (_m *DisplayManager) ShowInfo(message string) {
	_m.Called(message)
}

// DisplayManager_ShowInfo_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'ShowInfo'
type DisplayManager_ShowInfo_Call struct {
	*mock.Call
}

// ShowInfo is a helper method to define mock.On call
//   - message string
func (_e *DisplayManager_Expecter) ShowInfo(message interface{}) *DisplayManager_ShowInfo_Call {
	return &DisplayManager_ShowInfo_Call{Call: _e.mock.On("ShowInfo", message)}
}

func (_c *DisplayManager_ShowInfo_Call) Run(run func(message string)) *DisplayManager_ShowInfo_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(string))
	})
	return _c
}

func (_c *DisplayManager_ShowInfo_Call) Return() *DisplayManager_ShowInfo_Call {
	_c.Call.Return()
	return _c
}

func (_c *DisplayManager_ShowInfo_Call) RunAndReturn(run func(string)) *DisplayManager_ShowInfo_Call {
	_c.Call.Return(run)
	return _c
}

// ShowNumberedList provides a mock function with given fields: items
func (_m *DisplayManager) ShowNumberedList(items [][2]string) {
	_m.Called(items)
}

// DisplayManager_ShowNumberedList_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'ShowNumberedList'
type DisplayManager_ShowNumberedList_Call struct {
	*mock.Call
}

// ShowNumberedList is a helper method to define mock.On call
//   - items [][2]string
func (_e *DisplayManager_Expecter) ShowNumberedList(items interface{}) *DisplayManager_ShowNumberedList_Call {
	return &DisplayManager_ShowNumberedList_Call{Call: _e.mock.On("ShowNumberedList", items)}
}

func (_c *DisplayManager_ShowNumberedList_Call) Run(run func(items [][2]string)) *DisplayManager_ShowNumberedList_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].([][2]string))
	})
	return _c
}

func (_c *DisplayManager_ShowNumberedList_Call) Return() *DisplayManager_ShowNumberedList_Call {
	_c.Call.Return()
	return _c
}

func (_c *DisplayManager_ShowNumberedList_Call) RunAndReturn(run func([][2]string)) *DisplayManager_ShowNumberedList_Call {
	_c.Call.Return(run)
	return _c
}

// ShowSection provides a mock function with given fields: title, content, opts
func (_m *DisplayManager) ShowSection(title string, content string, opts map[string]string) {
	_m.Called(title, content, opts)
}

// DisplayManager_ShowSection_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'ShowSection'
type DisplayManager_ShowSection_Call struct {
	*mock.Call
}

// ShowSection is a helper method to define mock.On call
//   - title string
//   - content string
//   - opts map[string]string
func (_e *DisplayManager_Expecter) ShowSection(title interface{}, content interface{}, opts interface{}) *DisplayManager_ShowSection_Call {
	return &DisplayManager_ShowSection_Call{Call: _e.mock.On("ShowSection", title, content, opts)}
}

func (_c *DisplayManager_ShowSection_Call) Run(run func(title string, content string, opts map[string]string)) *DisplayManager_ShowSection_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(string), args[1].(string), args[2].(map[string]string))
	})
	return _c
}

func (_c *DisplayManager_ShowSection_Call) Return() *DisplayManager_ShowSection_Call {
	_c.Call.Return()
	return _c
}

func (_c *DisplayManager_ShowSection_Call) RunAndReturn(run func(string, string, map[string]string)) *DisplayManager_ShowSection_Call {
	_c.Call.Return(run)
	return _c
}

// ShowSuccess provides a mock function with given fields: message
func (_m *DisplayManager) ShowSuccess(message string) {
	_m.Called(message)
}

// DisplayManager_ShowSuccess_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'ShowSuccess'
type DisplayManager_ShowSuccess_Call struct {
	*mock.Call
}

// ShowSuccess is a helper method to define mock.On call
//   - message string
func (_e *DisplayManager_Expecter) ShowSuccess(message interface{}) *DisplayManager_ShowSuccess_Call {
	return &DisplayManager_ShowSuccess_Call{Call: _e.mock.On("ShowSuccess", message)}
}

func (_c *DisplayManager_ShowSuccess_Call) Run(run func(message string)) *DisplayManager_ShowSuccess_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(string))
	})
	return _c
}

func (_c *DisplayManager_ShowSuccess_Call) Return() *DisplayManager_ShowSuccess_Call {
	_c.Call.Return()
	return _c
}

func (_c *DisplayManager_ShowSuccess_Call) RunAndReturn(run func(string)) *DisplayManager_ShowSuccess_Call {
	_c.Call.Return(run)
	return _c
}

// ShowWarning provides a mock function with given fields: message
func (_m *DisplayManager) ShowWarning(message string) {
	_m.Called(message)
}

// DisplayManager_ShowWarning_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'ShowWarning'
type DisplayManager_ShowWarning_Call struct {
	*mock.Call
}

// ShowWarning is a helper method to define mock.On call
//   - message string
func (_e *DisplayManager_Expecter) ShowWarning(message interface{}) *DisplayManager_ShowWarning_Call {
	return &DisplayManager_ShowWarning_Call{Call: _e.mock.On("ShowWarning", message)}
}

func (_c *DisplayManager_ShowWarning_Call) Run(run func(message string)) *DisplayManager_ShowWarning_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(string))
	})
	return _c
}

func (_c *DisplayManager_ShowWarning_Call) Return() *DisplayManager_ShowWarning_Call {
	_c.Call.Return()
	return _c
}

func (_c *DisplayManager_ShowWarning_Call) RunAndReturn(run func(string)) *DisplayManager_ShowWarning_Call {
	_c.Call.Return(run)
	return _c
}

// StartSpinner provides a mock function with given fields: message
func (_m *DisplayManager) StartSpinner(message string) {
	_m.Called(message)
}

// DisplayManager_StartSpinner_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'StartSpinner'
type DisplayManager_StartSpinner_Call struct {
	*mock.Call
}

// StartSpinner is a helper method to define mock.On call
//   - message string
func (_e *DisplayManager_Expecter) StartSpinner(message interface{}) *DisplayManager_StartSpinner_Call {
	return &DisplayManager_StartSpinner_Call{Call: _e.mock.On("StartSpinner", message)}
}

func (_c *DisplayManager_StartSpinner_Call) Run(run func(message string)) *DisplayManager_StartSpinner_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(string))
	})
	return _c
}

func (_c *DisplayManager_StartSpinner_Call) Return() *DisplayManager_StartSpinner_Call {
	_c.Call.Return()
	return _c
}

func (_c *DisplayManager_StartSpinner_Call) RunAndReturn(run func(string)) *DisplayManager_StartSpinner_Call {
	_c.Call.Return(run)
	return _c
}

// StopSpinner provides a mock function with given fields:
func (_m *DisplayManager) StopSpinner() {
	_m.Called()
}

// DisplayManager_StopSpinner_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'StopSpinner'
type DisplayManager_StopSpinner_Call struct {
	*mock.Call
}

// StopSpinner is a helper method to define mock.On call
func (_e *DisplayManager_Expecter) StopSpinner() *DisplayManager_StopSpinner_Call {
	return &DisplayManager_StopSpinner_Call{Call: _e.mock.On("StopSpinner")}
}

func (_c *DisplayManager_StopSpinner_Call) Run(run func()) *DisplayManager_StopSpinner_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *DisplayManager_StopSpinner_Call) Return() *DisplayManager_StopSpinner_Call {
	_c.Call.Return()
	return _c
}

func (_c *DisplayManager_StopSpinner_Call) RunAndReturn(run func()) *DisplayManager_StopSpinner_Call {
	_c.Call.Return(run)
	return _c
}

// NewDisplayManager creates a new instance of DisplayManager. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewDisplayManager(t interface {
	mock.TestingT
	Cleanup(func())
}) *DisplayManager {
	mock := &DisplayManager{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
