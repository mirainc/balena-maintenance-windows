package main

import (
	"errors"
	"fmt"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gofrs/flock"
	"github.com/levigross/grequests"
)

var BALENA_API_KEY = os.Getenv("BALENA_API_KEY")
var BALENA_DEVICE_UUID = os.Getenv("BALENA_DEVICE_UUID")
var BALENA_API_BASE_URL = "https://api.balena-cloud.com"
var MAINTENANCE_WINDOW_TAG_KEY = "MAINTENANCE_WINDOW"
var TIME_FORMAT = "15:04:05MST"

type BalenaDeviceTag struct {
	Id     int    `json:"id"`
	TagKey string `json:"tag_key"`
	Value  string `json:"value"`
}

type BalenaDeviceTagResponse struct {
	Data []BalenaDeviceTag `json:"d"`
}

func UrlEncoded(str string) (string, error) {
	u, err := url.Parse(str)
	if err != nil {
		return "", err
	}
	return u.String(), nil
}

func getLockfileLocation() (string, error) {
	result := os.Getenv("LOCKFILE_LOCATION")
	if result == "" {
		result = "/tmp/balena"
	}
	err := os.MkdirAll(result, os.ModePerm)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s/updates.lock", result), nil
}

func getCheckInterval() time.Duration {
	interval := 10 * 60 * time.Second
	CHECK_INTERVAL_SECONDS := os.Getenv("CHECK_INTERVAL_SECONDS")
	if CHECK_INTERVAL_SECONDS != "" {
		parsedInterval, err := strconv.Atoi(CHECK_INTERVAL_SECONDS)
		if err != nil {
			fmt.Println("Failed to parse CHECK_INTERVAL_SECONDS, using default check interval.")
		} else {
			interval = time.Duration(parsedInterval) * time.Second
		}
	}
	return interval
}

func getMaintenanceWindow() (start *time.Time, end *time.Time, err error) {
	maintenanceWindowValue, err := getMaintenanceWindowTagValue()
	if err != nil {
		return nil, nil, err
	}

	return parseMaintenanceWindow(maintenanceWindowValue)
}

func parseMaintenanceWindow(value string) (*time.Time, *time.Time, error) {
	if value == "" {
		start := time.Date(0, 1, 1, 0, 0, 0, 0, time.UTC)
		end := time.Date(0, 1, 1, 23, 59, 59, 999999999, time.UTC)
		return &start, &end, nil
	}

	values := strings.Split(value, "_")
	if len(values) != 2 {
		return nil, nil, errors.New(fmt.Sprintf("Expected 2 timestamps, received %d. Tag value: %s", len(values), value))
	}

	start, err := time.Parse(TIME_FORMAT, values[0])
	if err != nil {
		return nil, nil, err
	}
	end, err := time.Parse(TIME_FORMAT, values[1])
	if err != nil {
		return nil, nil, err
	}

	if start.After(end) {
		return nil, nil, errors.New(fmt.Sprintf("Start time is after end time. Tag value: %s", value))
	}

	return &start, &end, err
}

func nowIsInMaintenanceWindow(start time.Time, end time.Time) bool {
	// Get the current time
	now := time.Now()
	dateAgnosticNow := time.Date(0, 1, 1, now.Hour(), now.Minute(), now.Second(), now.Nanosecond(), now.Location())

	fmt.Println("Start:", start.UTC())
	fmt.Println("End:", end.UTC())
	fmt.Println("Now:", dateAgnosticNow.UTC())

	return isInMaintenanceWindow(dateAgnosticNow.UTC(), start.UTC(), end.UTC())
}

func isInMaintenanceWindow(test time.Time, start time.Time, end time.Time) bool {
	return test.After(start) && test.Before(end)
}

// Fetch tag value from Balena
func getMaintenanceWindowTagValue() (string, error) {
	options := &grequests.RequestOptions{
		Headers: map[string]string{
			"Content-Type":  "application/json",
			"Authorization": fmt.Sprintf("Bearer %s", BALENA_API_KEY),
		},
		RequestTimeout: time.Duration(5) * time.Second,
	}
	filters, err := UrlEncoded(fmt.Sprintf("$filter=device/uuid eq '%s'&$filter=tag_key eq '%s'", BALENA_DEVICE_UUID, MAINTENANCE_WINDOW_TAG_KEY))
	if err != nil {
		return "", err
	}

	url := fmt.Sprintf("%s/v4/device_tag?%s", BALENA_API_BASE_URL, filters)
	resp, err := grequests.Get(url, options)
	if err != nil {
		return "", err
	}
	if !resp.Ok {
		message := fmt.Sprintf("Request failed with response status code: %s", resp.StatusCode)
		return "", errors.New(message)
	}

	tags := new(BalenaDeviceTagResponse)
	err = resp.JSON(tags)
	if err != nil {
		return "", err
	}

	if len(tags.Data) == 0 {
		fmt.Println("No maintenance window set, default to any time.")
		return "", nil
	} else {
		window := tags.Data[0].Value
		fmt.Println("Maintenance window found: [", window, "]")
		return window, nil
	}
}

func loopIteration(lock *flock.Flock) {
	start, end, err := getMaintenanceWindow()
	if err != nil {
		fmt.Println("Failed to get maintenance window:", err.Error())
	} else {
		shouldLock := !nowIsInMaintenanceWindow(*start, *end)

		if shouldLock {
			fmt.Println("Not in maintenance window, taking lock...")
			locked, err := lock.TryLock()
			if err != nil {
				fmt.Println("Failed to take lock:", err.Error())
			}
			if locked {
				fmt.Println("Lock taken successfully.")
			} else {
				fmt.Println("Failed to take lock - locked by another thread.")
			}
		} else {
			fmt.Println("In maintenance window, unlocking...")
			err := lock.Unlock()
			if err != nil {
				fmt.Println("Failed to unlock:", err.Error())
			} else {
				fmt.Println("Unlocked successfully.")
			}
		}
	}
}

func loop(lock *flock.Flock, interval time.Duration) {
	for {
		loopIteration(lock)
		fmt.Println("Waiting", interval.Seconds(), "seconds...")
		time.Sleep(interval)
	}
}

func main() {
	// Get lock location and create lock.
	lockfileLocation, err := getLockfileLocation()
	if err != nil {
		fmt.Println("Failed to get lockfile location:", err.Error())
	}
	fmt.Println("Using lock location:", lockfileLocation)

	interval := getCheckInterval()
	fmt.Println("Check interval (seconds):", interval.Seconds())

	lock := flock.New(lockfileLocation)

	// Start check process.
	loop(lock, interval)

	fmt.Println("Exited.")
}
