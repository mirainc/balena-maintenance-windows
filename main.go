package main

import (
	"fmt"
	"os"
	"time"

	"github.com/dghubble/sling"
	"github.com/gofrs/flock"
)

var BALENA_API_KEY = os.Getenv("BALENA_API_KEY")
var BALENA_DEVICE_UUID = os.Getenv("BALENA_DEVICE_UUID")
var BALENA_API_BASE_URL = "https://api.balena-cloud.com/v4"
var MAINTENANCE_WINDOW_TAG_KEY = "MAINTENANCE_WINDOW"

type BalenaDeviceTag struct {
	Id     int    `json:"id"`
	TagKey string `json:"tag_key"`
	Value  string `json:"value"`
}

type BalenaDeviceTagResponse struct {
	Data []BalenaDeviceTag `json:"d"`
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

func getMaintenanceWindow(client *sling.Sling) (start *time.Time, end *time.Time, err error) {
	start = nil
	end = nil

	tags := new(BalenaDeviceTagResponse)
	url := fmt.Sprintf("device_tag?$filter=device/uuid eq '%s'&$filter=tag_key eq '%s'", BALENA_DEVICE_UUID, MAINTENANCE_WINDOW_TAG_KEY)
	_, err = client.New().Get(url).ReceiveSuccess(tags)
	if len(tags.Data) == 0 {
		fmt.Println("No maintenance window set, default to any time.")
		return nil, nil, nil
	} else {
		window := tags.Data[0].Value
		fmt.Println("Maintenance window found:", window)
	}
	return start, end, err
}

func makeBalenaClient() *sling.Sling {
	s := sling.New().
		Base(BALENA_API_BASE_URL).
		Set("Content-Type", "application/json").
		Set("Authorization", fmt.Sprintf("Bearer %s", BALENA_API_KEY))
	return s
}

func main() {
	// Get lock location and create lock.
	lockfileLocation, err := getLockfileLocation()
	if err != nil {
		fmt.Println("Failed to get lockfile location:", err.Error())
	}
	fmt.Println("Using lock location:", lockfileLocation)

	fileLock := flock.New(lockfileLocation)

	// Start check process.
	client := makeBalenaClient()

	start, end, err := getMaintenanceWindow(client)
	fmt.Println(start)
	fmt.Println(end)

	locked, err := fileLock.TryLock()
	if err != nil {
		// handle locking error
		fmt.Println("Failed to take lock:", err.Error())
	}

	if locked {
		// do work
		fmt.Println("Lock taken successfully.")

		fmt.Println("Unlocking...")
		err := fileLock.Unlock()
		if err != nil {
			fmt.Println("Failed to unlock:", err.Error())
		}
	}

	fmt.Println("Done.")
}
