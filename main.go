package main

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/mirainc/balena-maintenance-windows/balenaapi"
	"github.com/mirainc/balena-maintenance-windows/lockfile"
	"github.com/mirainc/balena-maintenance-windows/timeutils"
	log "github.com/sirupsen/logrus"
)

var BALENA_API_KEY = os.Getenv("BALENA_API_KEY")
var BALENA_DEVICE_UUID = os.Getenv("BALENA_DEVICE_UUID")
var MAINTENANCE_WINDOW_TAG_KEY = "MAINTENANCE_WINDOW"

var logger = log.WithFields(log.Fields{})

func initLogger() {
	log.SetFormatter(&log.JSONFormatter{})

	log.SetOutput(os.Stdout)

	logLevel := os.Getenv("LOG_LEVEL")
	if logLevel == "" {
		logLevel = "info"
	}
	level, err := log.ParseLevel(logLevel)
	if err != nil {
		log.SetLevel(log.InfoLevel)
	} else {
		log.SetLevel(level)
	}
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
			logger.Warn("Failed to parse CHECK_INTERVAL_SECONDS, using default check interval.")
		} else {
			interval = time.Duration(parsedInterval) * time.Second
		}
	}
	return interval
}

func loopIteration(lock *lockfile.Lockfile) {
	now := time.Now()

	maintenanceWindowValue, err := balenaapi.GetTagValue(BALENA_API_KEY, BALENA_DEVICE_UUID, MAINTENANCE_WINDOW_TAG_KEY)
	if err != nil {
		logger.Warn("Failed to get maintenance window tag: ", err.Error())
		return
	}

	inWindow, err := timeutils.IsInMaintenanceWindow(maintenanceWindowValue, now)
	if err != nil {
		logger.Error("Failed to parse maintenance window: ", err.Error())
	} else {
		if !inWindow {
			logger.Info("Not in maintenance window, taking lock...")
			err := lock.Lock()
			if err != nil {
				logger.Error("Failed to take lock:", err.Error())
			} else {
				logger.Info("Lock taken successfully.")
			}
		} else {
			logger.Info("In maintenance window, unlocking...")
			err := lock.Unlock()
			if err != nil {
				logger.Error("Failed to unlock: ", err.Error())
			} else {
				logger.Info("Unlocked successfully.")
			}
		}
	}
}

func loop(lock *lockfile.Lockfile, interval time.Duration) {
	for {
		loopIteration(lock)
		logger.Info("Waiting ", interval.Seconds(), " seconds...")
		time.Sleep(interval)
	}
}

func main() {
	initLogger()

	// Get lock location and create lock.
	lockfileLocation, err := getLockfileLocation()
	if err != nil {
		logger.Error("Failed to get lockfile location: ", err.Error())
		panic(err)
	}
	logger.Info("Using lock location: ", lockfileLocation)

	interval := getCheckInterval()
	logger.Info("Check interval (seconds): ", interval.Seconds())

	lock := lockfile.New(lockfileLocation)

	// Start check process.
	loop(lock, interval)

	logger.Info("Exited.")
}
