/**
 * This file is part of the raoptimus/db-migrator.go library
 *
 * @copyright Copyright (c) Evgeniy Urvantsev
 * @license https://github.com/raoptimus/db-migrator.go/blob/master/LICENSE.md
 * @link https://github.com/raoptimus/db-migrator.go
 */

package urfavecli

import (
	"context"
	"errors"
	"testing"

	"github.com/raoptimus/db-migrator.go/internal/application/handler"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/urfave/cli/v3"
)

// MockHandler is a mock type for the Handler interface
type MockHandler struct {
	mock.Mock
}

// Handle provides a mock function with given fields: cmd
func (m *MockHandler) Handle(cmd *handler.Command) error {
	ret := m.Called(cmd)

	if len(ret) == 0 {
		panic("no return value specified for Handle")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(*handler.Command) error); ok {
		r0 = rf(cmd)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// NewMockHandler creates a new instance of MockHandler
func NewMockHandler(t *testing.T) *MockHandler {
	mockHandler := &MockHandler{}
	mockHandler.Mock.Test(t)

	t.Cleanup(func() { mockHandler.AssertExpectations(t) })

	return mockHandler
}

func TestAdapt_Handle_NoArgs_Successfully(t *testing.T) {
	tests := []struct {
		name string
	}{
		{
			name: "empty args",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockHandler := NewMockHandler(t)

			mockHandler.On("Handle", mock.MatchedBy(func(cmd *handler.Command) bool {
				return cmd != nil && cmd.Context() != nil && !cmd.Args.Present()
			})).Return(nil)

			actionFunc := Adapt(mockHandler)
			require.NotNil(t, actionFunc)

			cliCmd := &cli.Command{
				Name:   "test",
				Action: actionFunc,
			}

			ctx := context.Background()
			err := cliCmd.Run(ctx, []string{"test"})
			require.NoError(t, err)
		})
	}
}

func TestAdapt_Handle_WithArgs_Successfully(t *testing.T) {
	tests := []struct {
		name         string
		args         []string
		expectedArgs []string
	}{
		{
			name:         "single arg",
			args:         []string{"arg1"},
			expectedArgs: []string{"arg1"},
		},
		{
			name:         "multiple args",
			args:         []string{"arg1", "arg2", "arg3"},
			expectedArgs: []string{"arg1", "arg2", "arg3"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockHandler := NewMockHandler(t)

			var capturedCmd *handler.Command
			mockHandler.On("Handle", mock.MatchedBy(func(cmd *handler.Command) bool {
				capturedCmd = cmd
				return cmd != nil
			})).Return(nil)

			actionFunc := Adapt(mockHandler)
			require.NotNil(t, actionFunc)

			cliCmd := &cli.Command{
				Name:   "test",
				Action: actionFunc,
			}

			ctx := context.Background()
			err := cliCmd.Run(ctx, append([]string{"test"}, tt.args...))
			require.NoError(t, err)

			require.NotNil(t, capturedCmd)
			require.Equal(t, tt.expectedArgs, capturedCmd.Args.Slice())
		})
	}
}

func TestAdapt_Handle_ContextPassed_Successfully(t *testing.T) {
	tests := []struct {
		name string
	}{
		{
			name: "context is passed correctly",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockHandler := NewMockHandler(t)

			type ctxKey string
			testKey := ctxKey("testKey")
			testValue := "testValue"

			var capturedCtx context.Context
			mockHandler.On("Handle", mock.MatchedBy(func(cmd *handler.Command) bool {
				capturedCtx = cmd.Context()
				return cmd != nil
			})).Return(nil)

			actionFunc := Adapt(mockHandler)
			require.NotNil(t, actionFunc)

			cliCmd := &cli.Command{
				Name:   "test",
				Action: actionFunc,
			}

			ctx := context.WithValue(context.Background(), testKey, testValue)
			err := cliCmd.Run(ctx, []string{"test"})
			require.NoError(t, err)

			require.NotNil(t, capturedCtx)
			require.Equal(t, testValue, capturedCtx.Value(testKey))
		})
	}
}

func TestAdapt_Handle_HandlerReturnsError_Failure(t *testing.T) {
	errHandler := errors.New("handler error")

	tests := []struct {
		name        string
		handlerErr  error
		expectedErr error
	}{
		{
			name:        "handler returns error",
			handlerErr:  errHandler,
			expectedErr: errHandler,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockHandler := NewMockHandler(t)

			mockHandler.On("Handle", mock.MatchedBy(func(cmd *handler.Command) bool {
				return cmd != nil
			})).Return(tt.handlerErr)

			actionFunc := Adapt(mockHandler)
			require.NotNil(t, actionFunc)

			cliCmd := &cli.Command{
				Name:   "test",
				Action: actionFunc,
			}

			ctx := context.Background()
			err := cliCmd.Run(ctx, []string{"test"})
			require.Error(t, err)
			require.ErrorIs(t, err, tt.expectedErr)
		})
	}
}

func TestAdapt_ReturnsNonNilActionFunc_Successfully(t *testing.T) {
	tests := []struct {
		name string
	}{
		{
			name: "returns non-nil action func",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockHandler := NewMockHandler(t)

			actionFunc := Adapt(mockHandler)
			require.NotNil(t, actionFunc)
		})
	}
}
