/**
 * This file is part of the raoptimus/db-migrator.go library
 *
 * @copyright Copyright (c) Evgeniy Urvantsev
 * @license https://github.com/raoptimus/db-migrator.go/blob/master/LICENSE.md
 * @link https://github.com/raoptimus/db-migrator.go
 */

package handler

import "context"

// Command represents a CLI command with its arguments and execution context.
type Command struct {
	Args Args
	ctx  context.Context
}

// Context returns the command's context.
// If no context is set, it returns context.Background().
func (c *Command) Context() context.Context {
	if c.ctx != nil {
		return c.ctx
	}

	return context.Background()
}

// WithContext creates a shallow copy of the command with a new context.
// It panics if the provided context is nil.
func (c *Command) WithContext(ctx context.Context) *Command {
	if ctx == nil {
		panic("nil context")
	}
	c2 := new(Command)
	*c2 = *c
	c2.ctx = ctx

	return c2
}

// Args provides access to command-line arguments.
//
//go:generate mockery
type Args interface {
	// First returns the first argument, or else a blank string
	First() string
	// Present checks if there are any arguments present
	Present() bool
	// Slice returns a copy of the internal slice
	Slice() []string
}

// Handler handles CLI commands.
//
//go:generate mockery
type Handler interface {
	Handle(cmd *Command) error
}
