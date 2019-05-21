package main

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/gofrs/flock"
	"github.com/mirainc/balena-maintenance-windows/balenaapi"
	"github.com/mirainc/balena-maintenance-windows/timeutils"
)

var BALENA_API_KEY = os.Getenv("BALENA_API_KEY")
var BALENA_DEVICE_UUID = os.Getenv("BALENA_DEVICE_UUID")
var TIMEZONE = os.Getenv("TIMEZONE")

var MAINTENANCE_WINDOW_TAG_KEY = "MAINTENANCE_WINDOW"

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

func loopIteration(lock *flock.Flock) {
	now := time.Now()

	maintenanceWindowValue, err := balenaapi.GetTagValue(BALENA_API_KEY, BALENA_DEVICE_UUID, MAINTENANCE_WINDOW_TAG_KEY)
	if err != nil {
		fmt.Println("Failed to get maintenance window tag:", err.Error())
		return
	}

	inWindow, err := timeutils.IsInMaintenanceWindow(maintenanceWindowValue, now)
	if err != nil {
		fmt.Println("Failed to parse maintenance window:", err.Error())
	} else {
		if !inWindow {
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
