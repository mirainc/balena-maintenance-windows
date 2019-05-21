package timeutils

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func Test_IsInMaintenanceWindow__UTC_Nighttime(t *testing.T) {
	// TZ: UTC
	location := time.UTC

	// Window: 10PM-11PM
	value := "22:00:00_23:00:00"

	// Now: 11PM
	now := time.Date(2018, 1, 1, 22, 30, 0, 0, location)

	inWindow, err := IsInMaintenanceWindow(value, now)
	assert.Nil(t, err)
	assert.True(t, inWindow)

	// Now: 11AM
	now = time.Date(2018, 1, 1, 11, 0, 0, 0, location)

	inWindow, err = IsInMaintenanceWindow(value, now)
	assert.Nil(t, err)
	assert.False(t, inWindow)
}

func Test_IsInMaintenanceWindow__UTC_Multiday(t *testing.T) {
	// TZ: UTC
	location := time.UTC

	// Window: 10PM-4AM
	value := "22:00:00_04:00:00"

	// Now: 11PM
	now := time.Date(2018, 1, 1, 22, 30, 0, 0, location)

	inWindow, err := IsInMaintenanceWindow(value, now)
	assert.Nil(t, err)
	assert.True(t, inWindow)

	// Now: 12AM
	now = time.Date(2018, 1, 1, 0, 0, 0, 0, location)

	inWindow, err = IsInMaintenanceWindow(value, now)
	assert.Nil(t, err)
	assert.True(t, inWindow)

	// Now: 2AM
	now = time.Date(2018, 1, 1, 2, 0, 0, 0, location)

	inWindow, err = IsInMaintenanceWindow(value, now)
	assert.Nil(t, err)
	assert.True(t, inWindow)

	// Now: 6AM
	now = time.Date(2018, 1, 1, 6, 0, 0, 0, location)

	inWindow, err = IsInMaintenanceWindow(value, now)
	assert.Nil(t, err)
	assert.False(t, inWindow)
}

func Test_IsInMaintenanceWindow__PDT_Nighttime(t *testing.T) {
	// TZ: PDT
	location, err := time.LoadLocation("America/Los_Angeles")
	assert.Nil(t, err)

	// Window: 10PM-11:00PM
	value := "22:00:00_23:00:00"

	// Now: 11PM
	now := time.Date(2018, 1, 1, 22, 30, 0, 0, location)

	inWindow, err := IsInMaintenanceWindow(value, now)
	assert.Nil(t, err)
	assert.True(t, inWindow)

	// Now: 11AM
	now = time.Date(2018, 1, 1, 11, 0, 0, 0, location)

	inWindow, err = IsInMaintenanceWindow(value, now)
	assert.Nil(t, err)
	assert.False(t, inWindow)
}

func Test_IsInMaintenanceWindow__PDT_Multiday(t *testing.T) {
	// TZ: PDT
	location, err := time.LoadLocation("America/Los_Angeles")
	assert.Nil(t, err)

	// Window: 10PM-4AM
	value := "22:00:00_04:00:00"

	// Now: 11PM
	now := time.Date(2018, 1, 1, 22, 30, 0, 0, location)

	inWindow, err := IsInMaintenanceWindow(value, now)
	assert.Nil(t, err)
	assert.True(t, inWindow)

	// Now: 12AM
	now = time.Date(2018, 1, 1, 0, 0, 0, 0, location)

	inWindow, err = IsInMaintenanceWindow(value, now)
	assert.Nil(t, err)
	assert.True(t, inWindow)

	// Now: 2AM
	now = time.Date(2018, 1, 1, 2, 0, 0, 0, location)

	inWindow, err = IsInMaintenanceWindow(value, now)
	assert.Nil(t, err)
	assert.True(t, inWindow)

	// Now: 6AM
	now = time.Date(2018, 1, 1, 6, 0, 0, 0, location)

	inWindow, err = IsInMaintenanceWindow(value, now)
	assert.Nil(t, err)
	assert.False(t, inWindow)
}

func Test_IsInMaintenanceWindow__PDT_DstTransition(t *testing.T) {
	// TZ: PDT
	location, err := time.LoadLocation("America/Los_Angeles")
	assert.Nil(t, err)

	// Window: 2AM-3AM
	value := "02:00:00_03:00:00"

	// Now: 2AM
	// DST started on 2018/3/11 2AM.
	// At this time clocks will jump forward one hour,
	// meaning there should be no "now" between 2-3AM.
	// If the check occurs exactly at 2AM it should still show within the window.
	// If the check occurs exactly at 3AM, it will be outside the window.
	now := time.Date(2018, 3, 11, 2, 00, 0, 0, location)

	inWindow, err := IsInMaintenanceWindow(value, now)
	assert.Nil(t, err)
	assert.True(t, inWindow)

	now = time.Date(2018, 3, 11, 3, 00, 0, 0, location)

	inWindow, err = IsInMaintenanceWindow(value, now)
	assert.Nil(t, err)
	assert.False(t, inWindow)

	// Now: 2AM
	// DST ended on 2018/11/4 2AM.
	// At this time clocks jump backward one hour,
	// meaning the 1-2AM hour repeats twice.
	// The check should still function correctly.
	now = time.Date(2018, 11, 4, 2, 00, 0, 0, location)

	inWindow, err = IsInMaintenanceWindow(value, now)
	assert.Nil(t, err)
	assert.True(t, inWindow)

	now = time.Date(2018, 11, 4, 3, 00, 0, 0, location)

	inWindow, err = IsInMaintenanceWindow(value, now)
	assert.Nil(t, err)
	assert.False(t, inWindow)
}
