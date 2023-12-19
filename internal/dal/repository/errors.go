package repository

import (
	"strings"
)

type DBError struct {
	Code          string
	Severity      string
	Message       string
	Details       string
	InternalQuery string
}

func (d *DBError) Error() string {
	var sb strings.Builder
	sb.WriteString("SQLSTATE[")
	sb.WriteString(d.Code)
	sb.WriteString("]: ")

	if d.Severity != "" {
		sb.WriteString(d.Severity)
		sb.WriteString(": ")
	}

	sb.WriteString(d.Message)
	sb.WriteString("\n")

	if d.InternalQuery != "" {
		sb.WriteString("The SQL being executed was: ")
		sb.WriteString(d.InternalQuery)
		sb.WriteString("\n")
	}

	if d.Details != "" {
		sb.WriteString("Details: ")
		sb.WriteString(d.Details)
		sb.WriteString("\n")
	}

	return strings.TrimRight(sb.String(), "\n")
}
