package timeutils

import (
	"errors"
	"fmt"
	"strings"
	"time"
)

var TIME_FORMAT = "2006/1/2T15:04:05"

func getTimeFromStringOnSpecificDay(time string, year int, month time.Month, day int) string {
	fullTime := fmt.Sprintf("%d/%d/%dT%s", year, month, day, time)
	return fullTime
}

func parseMaintenanceWindow(value string, now time.Time) (*time.Time, *time.Time, error) {
	location := now.Location()
	year := now.Year()
	month := now.Month()
	day := now.Day()

	if value == "" {
		start := time.Date(year, month, day, 0, 0, 0, 0, location)
		end := time.Date(year, month, day, 23, 59, 59, 999999999, location)
		return &start, &end, nil
	}

	values := strings.Split(value, "_")
	if len(values) != 2 {
		return nil, nil, errors.New(fmt.Sprintf("Expected 2 timestamps, received %d. Tag value: %s", len(values), value))
	}

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

func isInMaintenanceWindowSimple(now time.Time, start time.Time, end time.Time) bool {
	return (now.After(start) && now.Before(end)) || now.Equal(start)
}

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
		fmt.Println("Start of Day:", sodToday)
		fmt.Println("End:", end)
		fmt.Println("Start:", start)
		fmt.Println("End of Day:", eodToday)
		fmt.Println("Now:", now)
		return isInMaintenanceWindowSimple(now, start, eodToday) || isInMaintenanceWindowSimple(now, sodToday, end)
	} else {
		fmt.Println("Start:", start)
		fmt.Println("End:", end)
		fmt.Println("Now:", now)
		return isInMaintenanceWindowSimple(now, start, end)
	}
}

func IsInMaintenanceWindow(value string, now time.Time) (bool, error) {
	start, end, err := parseMaintenanceWindow(value, now)
	if err != nil {
		return false, err
	}
	return isInMaintenanceWindow(now, *start, *end), nil
}
