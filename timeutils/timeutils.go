package timeutils

import (
	"errors"
	"fmt"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
)

var TIME_FORMAT = "2006/1/2T15:04:05"
var logger = log.WithFields(log.Fields{"package": "timeutils"})

// Given a valid timestamp, e.g. "17:00:00", return a formatted time string for the passed in year/month/day.
func getTimeFromStringOnSpecificDay(time string, year int, month time.Month, day int) string {
	fullTime := fmt.Sprintf("%d/%d/%dT%s", year, month, day, time)
	return fullTime
}

// Parses a valid maintenance window string, e.g. "17:00:00_23:00:00"
// Returns start and end time, normalized to the current day.
// Returned end time be before start time.
func parseMaintenanceWindow(value string, now time.Time) (*time.Time, *time.Time, error) {
	location := now.Location()
	year := now.Year()
	month := now.Month()
	day := now.Day()

	// If the maintenance window value is empty, set the window to be the entire day.
	if value == "" {
		start := time.Date(year, month, day, 0, 0, 0, 0, location)
		end := time.Date(year, month, day, 23, 59, 59, 999999999, location)
		return &start, &end, nil
	}

	values := strings.Split(value, "_")
	if len(values) != 2 {
		return nil, nil, errors.New(fmt.Sprintf("Expected 2 timestamps, received %d. Tag value: %s", len(values), value))
	}

	// Convert the time to the current day, to account for local time changes (e.g. daylight savings time).
	startString := getTimeFromStringOnSpecificDay(values[0], year, month, day)
	endString := getTimeFromStringOnSpecificDay(values[1], year, month, day)

	start, err := time.ParseInLocation(TIME_FORMAT, startString, location)
	if err != nil {
		return nil, nil, err
	}
	end, err := time.ParseInLocation(TIME_FORMAT, endString, location)
	if err != nil {
		return nil, nil, err
	}

	return &start, &end, err
}

// Evaluate maintenance window assuming start is before end time.
func isInMaintenanceWindowSimple(now time.Time, start time.Time, end time.Time) bool {
	return (now.After(start) && now.Before(end)) || now.Equal(start)
}

// Evaluate maintenance window, handling cases where end is before start time.
func isInMaintenanceWindow(now time.Time, start time.Time, end time.Time) bool {
	location := now.Location()
	year := now.Year()
	month := now.Month()
	day := now.Day()

	if start.After(end) {
		// Since start and end times are calibrated to the same day, this can
		// happen if the time window expects to cross midnight, e.g. 22:00-02:00.
		// Handle this by creating two time windows.
		sodToday := time.Date(year, month, day, 0, 0, 0, 0, location)
		eodToday := time.Date(year, month, day, 23, 59, 59, 999999999, location)
		logger.Debug("[ ", sodToday, " - ", end, " + ", start, " - ", eodToday, " ]")
		logger.Debug("Now: ", now)
		return isInMaintenanceWindowSimple(now, start, eodToday) || isInMaintenanceWindowSimple(now, sodToday, end)
	} else {
		logger.Debug("[ ", start, " - ", end, " ]")
		logger.Debug("Now:", now)
		return isInMaintenanceWindowSimple(now, start, end)
	}
}

// Evaluate maintenance window given a valid maintenance window string, e.g. "17:00:00_23:00:00".
func IsInMaintenanceWindow(value string, now time.Time) (bool, error) {
	start, end, err := parseMaintenanceWindow(value, now)
	if err != nil {
		return false, err
	}
	return isInMaintenanceWindow(now, *start, *end), nil
}
