/**
 * This file is part of the raoptimus/db-migrator.go library
 *
 * @copyright Copyright (c) Evgeniy Urvantsev
 * @license https://github.com/raoptimus/db-migrator.go/blob/master/LICENSE.md
 * @link https://github.com/raoptimus/db-migrator.go
 */

package console

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

var (
	Black   = color("\033[1;30m%s\033[0m")
	Red     = color("\033[1;31m%s\033[0m")
	Green   = color("\033[1;32m%s\033[0m")
	Yellow  = color("\033[1;33m%s\033[0m")
	Purple  = color("\033[1;34m%s\033[0m")
	Magenta = color("\033[1;35m%s\033[0m")
	Teal    = color("\033[1;36m%s\033[0m")
	White   = color("\033[1;37m%s\033[0m")
)

type Console struct{}

var Std = New()

func New() *Console {
	return &Console{}
}

func (c *Console) Confirm(s string) bool {
	reader := bufio.NewReader(os.Stdin)

	for {
		c.Infof("%s [y/n]: ", s)

		response, err := reader.ReadString('\n')
		if err != nil {
			c.Fatal(err)
		}

		response = strings.ToLower(strings.TrimSpace(response))

		if response == "y" || response == "yes" {
			return true
		} else if response == "n" || response == "no" {
			return false
		}
	}
}

func (c *Console) Info(message string) {
	fmt.Print(Black(message))
}
func (c *Console) InfoLn(message string) {
	fmt.Println(Black(message))
}
func (c *Console) Infof(message string, a ...any) {
	fmt.Printf(Black(message), a...)
}

func (c *Console) Success(message string) {
	fmt.Print(Green(message))
}
func (c *Console) SuccessLn(message string) {
	fmt.Println(Green(message))
}
func (c *Console) Successf(message string, a ...any) {
	fmt.Printf(Green(message), a...)
}

func (c *Console) Warn(message string) {
	fmt.Print(Yellow(message))
}
func (c *Console) WarnLn(message string) {
	fmt.Println(Yellow(message))
}
func (c *Console) Warnf(message string, a ...any) {
	fmt.Printf(Yellow(message), a...)
}

func (c *Console) Error(message string) {
	fmt.Print(Red(message))
}
func (c *Console) ErrorLn(message string) {
	fmt.Println(Red(message))
}
func (c *Console) Errorf(message string, a ...any) {
	fmt.Printf(Red(message), a...)
}

func (c *Console) Fatal(err error) {
	c.Errorf("Exception: %v", err)
	os.Exit(1)
}

func (c *Console) NumberPlural(count int, one, many string) string {
	if count > 1 {
		return many
	}

	return one
}

func color(colorString string) func(...interface{}) string {
	sprint := func(args ...interface{}) string {
		return fmt.Sprintf(colorString, fmt.Sprint(args...))
	}

	return sprint
}
