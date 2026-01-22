/**
 * This file is part of the raoptimus/db-migrator.go library
 *
 * @copyright Copyright (c) Evgeniy Urvantsev
 * @license https://github.com/raoptimus/db-migrator.go/blob/master/LICENSE.md
 * @link https://github.com/raoptimus/db-migrator.go
 */

package timex

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestNew_ReturnsTimeInstance_Successfully(t *testing.T) {
	tests := []struct {
		name         string
		expectedTime time.Time
	}{
		{
			name:         "returns time with zero value",
			expectedTime: time.Time{},
		},
		{
			name:         "returns time with specific date",
			expectedTime: time.Date(2024, 1, 15, 10, 30, 45, 0, time.UTC),
		},
		{
			name:         "returns time with different timezone",
			expectedTime: time.Date(2023, 12, 31, 23, 59, 59, 999999999, time.FixedZone("CET", 3600)),
		},
		{
			name:         "returns time at unix epoch",
			expectedTime: time.Unix(0, 0).UTC(),
		},
		{
			name:         "returns time far in the future",
			expectedTime: time.Date(2100, 6, 15, 12, 0, 0, 0, time.UTC),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			nowFunc := func() time.Time {
				return tt.expectedTime
			}

			result := New(nowFunc)

			require.NotNil(t, result)
			require.Equal(t, tt.expectedTime, result.Now())
		})
	}
}

func TestStdTime_Now_ReturnsConfiguredTime_Successfully(t *testing.T) {
	tests := []struct {
		name         string
		expectedTime time.Time
	}{
		{
			name:         "returns configured time with nanoseconds",
			expectedTime: time.Date(2024, 3, 20, 14, 25, 36, 123456789, time.UTC),
		},
		{
			name:         "returns configured time at midnight",
			expectedTime: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		},
		{
			name:         "returns configured time at end of day",
			expectedTime: time.Date(2024, 12, 31, 23, 59, 59, 999999999, time.UTC),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			timeInstance := New(func() time.Time {
				return tt.expectedTime
			})

			result := timeInstance.Now()

			require.Equal(t, tt.expectedTime, result)
		})
	}
}

func TestStdTime_Now_CallsNowFuncEachTime_Successfully(t *testing.T) {
	callCount := 0
	times := []time.Time{
		time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC),
		time.Date(2024, 1, 3, 0, 0, 0, 0, time.UTC),
	}

	timeInstance := New(func() time.Time {
		idx := callCount
		callCount++
		if idx < len(times) {
			return times[idx]
		}
		return time.Time{}
	})

	// First call
	result1 := timeInstance.Now()
	require.Equal(t, times[0], result1)

	// Second call
	result2 := timeInstance.Now()
	require.Equal(t, times[1], result2)

	// Third call
	result3 := timeInstance.Now()
	require.Equal(t, times[2], result3)

	require.Equal(t, 3, callCount)
}

func TestStdTime_GlobalVariable_IsNotNil_Successfully(t *testing.T) {
	require.NotNil(t, StdTime)
}

func TestStdTime_GlobalVariable_ReturnsCurrentTime_Successfully(t *testing.T) {
	before := time.Now()
	result := StdTime.Now()
	after := time.Now()

	require.False(t, result.Before(before), "StdTime.Now() should not return time before the test started")
	require.False(t, result.After(after), "StdTime.Now() should not return time after the test ended")
}
