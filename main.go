package main

import (
	"fmt"
	"github.com/gofrs/flock"
	"os"
)

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

func main() {
	lockfileLocation, err := getLockfileLocation()
	if err != nil {
		fmt.Println("Failed to get lockfile location:", err.Error())
	}
	fmt.Println("Using lock location:", lockfileLocation)

	fileLock := flock.New(lockfileLocation)

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
