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
var TIMEZONE = os.Getenv("TIMEZONE")
var BALENA_API_BASE_URL = "https://api.balena-cloud.com"
var MAINTENANCE_WINDOW_TAG_KEY = "MAINTENANCE_WINDOW"
var TIME_FORMAT = "2006/1/2T15:04:05"

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

func getMaintenanceWindow(now time.Time) (start *time.Time, end *time.Time, err error) {
	maintenanceWindowValue, err := getMaintenanceWindowTagValue()
	if err != nil {
		return nil, nil, err
	}

	return parseMaintenanceWindow(maintenanceWindowValue, now)
}

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
		return (now.After(start) && now.Before(eodToday)) || (now.After(sodToday) && now.Before(end))
	} else {
		fmt.Println("Start:", start)
		fmt.Println("End:", end)
		fmt.Println("Now:", now)
		return now.After(start) && now.Before(end)
	}
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
	now := time.Now()
	start, end, err := getMaintenanceWindow(now)
	if err != nil {
		fmt.Println("Failed to get maintenance window:", err.Error())
	} else {
		shouldLock := !isInMaintenanceWindow(now, *start, *end)

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
