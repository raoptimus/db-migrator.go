/**
 * This file is part of the raoptimus/db-migrator.go library
 *
 * @copyright Copyright (c) Evgeniy Urvantsev
 * @license https://github.com/raoptimus/db-migrator.go/blob/master/LICENSE.md
 * @link https://github.com/raoptimus/db-migrator.go
 */

package console

import (
	"fmt"
	"os"
)

var (
	Black   = Color("\033[1;30m%s\033[0m")
	Red     = Color("\033[1;31m%s\033[0m")
	Green   = Color("\033[1;32m%s\033[0m")
	Yellow  = Color("\033[1;33m%s\033[0m")
	Purple  = Color("\033[1;34m%s\033[0m")
	Magenta = Color("\033[1;35m%s\033[0m")
	Teal    = Color("\033[1;36m%s\033[0m")
	White   = Color("\033[1;37m%s\033[0m")
)

func Info(message string) {
	fmt.Print(Black(message))
}
func InfoLn(message string) {
	fmt.Println(Black(message))
}
func Infof(message string, a ...any) {
	fmt.Printf(Black(message), a...)
}

func Success(message string) {
	fmt.Print(Green(message))
}
func SuccessLn(message string) {
	fmt.Println(Green(message))
}
func Successf(message string, a ...any) {
	fmt.Printf(Green(message), a...)
}

func Warn(message string) {
	fmt.Print(Yellow(message))
}
func WarnLn(message string) {
	fmt.Println(Yellow(message))
}
func Warnf(message string, a ...any) {
	fmt.Printf(Yellow(message), a...)
}

func Error(message string) {
	fmt.Print(Red(message))
}
func ErrorLn(message string) {
	fmt.Println(Red(message))
}
func Errorf(message string, a ...any) {
	fmt.Printf(Red(message), a...)
}

func Fatal(err error) {
	Errorf("Exception: %v", err)
	os.Exit(1)
}

func Color(colorString string) func(...interface{}) string {
	sprint := func(args ...interface{}) string {
		return fmt.Sprintf(colorString, fmt.Sprint(args...))
	}

	return sprint
}

func NumberPlural(c int, one, many string) string {
	if c > 1 {
		return many
	}

	return one
}
