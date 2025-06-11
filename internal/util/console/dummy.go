/**
 * This file is part of the raoptimus/db-migrator.go library
 *
 * @copyright Copyright (c) Evgeniy Urvantsev
 * @license https://github.com/raoptimus/db-migrator.go/blob/master/LICENSE.md
 * @link https://github.com/raoptimus/db-migrator.go
 */

package console

type Dummy struct {
	confirm bool
}

func NewDummy(confirm bool) *Dummy {
	return &Dummy{confirm: confirm}
}

func (c *Dummy) Confirm(s string) bool {
	return c.confirm
}

func (c *Dummy) Info(message string) {
}
func (c *Dummy) InfoLn(message string) {
}
func (c *Dummy) Infof(message string, a ...any) {
}
func (c *Dummy) Success(message string) {
}
func (c *Dummy) SuccessLn(message string) {
}
func (c *Dummy) Successf(message string, a ...any) {
}
func (c *Dummy) Warn(message string) {
}
func (c *Dummy) WarnLn(message string) {
}
func (c *Dummy) Warnf(message string, a ...any) {
}
func (c *Dummy) Error(message string) {
}
func (c *Dummy) ErrorLn(message string) {
}
func (c *Dummy) Errorf(message string, a ...any) {
}
func (c *Dummy) Fatal(err error) {
}

func (c *Dummy) NumberPlural(count int, one, many string) string {
	if count > 1 {
		return many
	}

	return one
}
