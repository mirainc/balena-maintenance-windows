# balena-maintenance-windows

Maintenance window support for balena devices.

## Design

This app is meant to be deployed as a container on balena devices. It will poll the balena API every 10 minutes to see if it is within a maintenance window, based on the `MAINTENANCE_WINDOW` tag assigned to that device. If there is no tag set, it assumes the device is always available for updates.

It uses the balena [Application Update Locking](https://www.balena.io/docs/learn/deploy/release-strategy/update-locking/) feature to block updates when not in the maintenance window.

## Parameters

balena internal env vars used:
```bash
BALENA_API_KEY
BALENA_DEVICE_UUID
```

Required env vars:
```bash
TIMEZONE=America/Los_Angeles
```

Optional env vars respected:
```bash
LOG_LEVEL=panic|fatal|error|warn|info|debug|trace
LOCKFILE_LOCATION=/tmp/balena
CHECK_INTERVAL_SECONDS=60
```

## Development

Run:
```
make
```
to run the app Dockerized.

Run:
```
make run-local
```
to build and run the Go application locally.
