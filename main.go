package main

import (
	"fmt"
	"net/url"
	"os"
	"time"

	// "github.com/dghubble/sling"
	"github.com/gofrs/flock"
	"github.com/levigross/grequests"
)

var BALENA_API_KEY = os.Getenv("BALENA_API_KEY")
var BALENA_DEVICE_UUID = os.Getenv("BALENA_DEVICE_UUID")
var BALENA_API_BASE_URL = "https://api.balena-cloud.com"
var MAINTENANCE_WINDOW_TAG_KEY = "MAINTENANCE_WINDOW"

type BalenaDeviceTag struct {
	Id     int    `json:"id"`
	TagKey string `json:"tag_key"`
	Value  string `json:"value"`
}

type BalenaDeviceTagResponse struct {
	Data []BalenaDeviceTag `json:"d"`
}

type BalenaFilterParams struct {
	Filter []string `url:"$filter"`
}

func (*BalenaFilterParams) EncodeValues(key string, v *url.Values) error {
	return nil
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

func getMaintenanceWindow() (start *time.Time, end *time.Time, err error) {
	start = nil
	end = nil

	resp, err := makeRequest()
	fmt.Println("response status:", resp.StatusCode)
	// fmt.Println(resp.Request.URL)
	tags := new(BalenaDeviceTagResponse)
	err = resp.JSON(tags)
	if err != nil {
		return nil, nil, err
	}

	if err != nil {
		fmt.Println("Failed to check tags from balena:", err.Error())
	}
	if len(tags.Data) == 0 {
		fmt.Println("No maintenance window set, default to any time.")
		return nil, nil, nil
	} else {
		window := tags.Data[0].Value
		fmt.Println("Maintenance window found:", window)
	}
	return start, end, err
}

func UrlEncoded(str string) (string, error) {
	u, err := url.Parse(str)
	if err != nil {
		return "", err
	}
	return u.String(), nil
}

// func makeBalenaClient() *sling.Sling {
// 	var httpClient = &http.Client{}
// 	s := sling.New().
// 		Base(BALENA_API_BASE_URL).
// 		Client(httpClient).
// 		Set("Content-Type", "application/json").
// 		Set("Authorization", fmt.Sprintf("Bearer %s", BALENA_API_KEY))
// 	return s
// }

func makeRequest() (*grequests.Response, error) {
	options := &grequests.RequestOptions{
		Headers: map[string]string{
			"Content-Type":  "application/json",
			"Authorization": fmt.Sprintf("Bearer %s", BALENA_API_KEY),
		},
		RequestTimeout: time.Duration(5) * time.Second,
	}
	url := fmt.Sprintf("%s/v4/device_tag?$filter=device/uuid%%20eq%%20'%s'&$filter=tag_key%%20eq%%20'%s'", BALENA_API_BASE_URL, BALENA_DEVICE_UUID, MAINTENANCE_WINDOW_TAG_KEY)
	finalUrl, err := UrlEncoded(url)
	fmt.Println(finalUrl)
	if err != nil {
		return nil, err
	}
	return grequests.Get(finalUrl, options)
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
	start, end, err := getMaintenanceWindow()
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
