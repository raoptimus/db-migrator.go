/**
 * This file is part of the raoptimus/db-migrator.go library
 *
 * @copyright Copyright (c) Evgeniy Urvantsev
 * @license https://github.com/raoptimus/db-migrator.go/blob/master/LICENSE.md
 * @link https://github.com/raoptimus/db-migrator.go
 */

package handler

import (
	"fmt"
	"regexp"
	"strconv"
	"time"

	"github.com/pkg/errors"

	"github.com/raoptimus/db-migrator.go/internal/domain/model"
)

const (
	all             = "all"
	minLimit        = 1
	timestampLength = 13 // YYMMDD_HHMMSS format length
)

// ErrArgumentMustBeGreaterThanZero is returned when a step argument is less than or equal to zero.
var ErrArgumentMustBeGreaterThanZero = errors.New("the step argument must be greater than 0")

func stepOrDefault(cmd *Command, defaults int) (int, error) {
	if !cmd.Args.Present() {
		return defaults, nil
	}

	value := cmd.Args.First()

	switch value {
	case "":
		return defaults, nil
	case all:
		return 0, nil
	default:
		i, err := strconv.Atoi(value)
		if err != nil {
			return -1, fmt.Errorf("the step argument %s is not valid", value)
		}

		if i < 1 {
			return -1, ErrArgumentMustBeGreaterThanZero
		}

		return i, nil
	}
}

// ErrTargetVersionRequired is returned when target version argument is missing.
var ErrTargetVersionRequired = errors.New("target version is required")

// extractTimestamp extracts the timestamp part (YYMMDD_HHMMSS) from a full migration version.
// Examples:
//   - "251002_184510" -> "251002_184510"
//   - "251002_184510_change_scheme" -> "251002_184510"
func extractTimestamp(version string) string {
	if len(version) < timestampLength {
		return version
	}
	// Timestamp format is always timestampLength characters: YYMMDD_HHMMSS
	return version[:timestampLength]
}

// parseTargetVersion extracts and normalizes a migration version from various input formats.
// Supported formats:
//   - Timestamp: "150101_185401"
//   - Full name: "150101_185401_create_news_table"
//   - DateTime: "2015-01-01 18:54:01"
//   - UNIX timestamp: "1392853618"
//
// Returns the normalized version in YYMMDD_HHMMSS format.
func parseTargetVersion(input string) (string, error) {
	if input == "" {
		return "", ErrTargetVersionRequired
	}

	// Case 1: Already in timestamp format (YYMMDD_HHMMSS)
	timestampPattern := regexp.MustCompile(`^\d{6}_\d{6}$`)
	if timestampPattern.MatchString(input) {
		return input, nil
	}

	// Case 2: Full migration name (YYMMDD_HHMMSS_name)
	// Extract first 13 characters: 6 digits + underscore + 6 digits
	fullNamePattern := regexp.MustCompile(`^(\d{6}_\d{6})_.+$`)
	if matches := fullNamePattern.FindStringSubmatch(input); matches != nil {
		return matches[1], nil
	}

	// Case 3: DateTime string "2015-01-01 18:54:01"
	if dt, err := time.Parse("2006-01-02 15:04:05", input); err == nil {
		return dt.Format("060102_150405"), nil
	}

	// Case 4: UNIX timestamp "1392853618" (must be at least 9 digits)
	// Validate: minimum timestamp is 946684800 (2000-01-01), which has 10 digits
	// But we allow 9+ digits to handle older timestamps
	if len(input) >= 9 && len(input) <= 10 {
		if timestamp, err := strconv.ParseInt(input, 10, 64); err == nil {
			dt := time.Unix(timestamp, 0).UTC()
			return dt.Format("060102_150405"), nil
		}
	}

	return "", fmt.Errorf("invalid version format: %s", input)
}

// applyMigrations applies a list of migrations.
// Returns number of applied migrations and error if any occurred.
func applyMigrations(
	cmd *Command,
	svc MigrationService,
	presenter Presenter,
	fileNameBuilder FileNameBuilder,
	migrations model.Migrations,
) (int, error) {
	applied := 0
	for i := range migrations {
		migration := &migrations[i]
		fileName, safely := fileNameBuilder.Up(migration.Version, false)

		if err := svc.ApplyFile(cmd.Context(), migration, fileName, safely); err != nil {
			presenter.ShowUpgradeError(applied, len(migrations))
			return applied, err
		}

		applied++
	}

	return applied, nil
}

// revertMigrations reverts a list of migrations.
// Returns number of reverted migrations and error if any occurred.
func revertMigrations(
	cmd *Command,
	svc MigrationService,
	presenter Presenter,
	fileNameBuilder FileNameBuilder,
	migrations model.Migrations,
) (int, error) {
	reverted := 0
	for i := range migrations {
		migration := &migrations[i]
		fileName, safely := fileNameBuilder.Down(migration.Version, false)

		if err := svc.RevertFile(cmd.Context(), migration, fileName, safely); err != nil {
			presenter.ShowDowngradeError(reverted, len(migrations))
			return reverted, err
		}

		reverted++
	}

	return reverted, nil
}
