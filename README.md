# balena-maintenance-windows

Maintenance window support for balena devices.

## Design

This app is meant to be deployed as a container on balena devices. It will poll the balena API every 10 minutes to see if it is within a maintenance window, based on the `MAINTENANCE_WINDOW` tag assigned to that device. If there is no tag set, it assumes the device is always available for updates.

It uses the balena [Application Update Locking](https://www.balena.io/docs/learn/deploy/release-strategy/update-locking/) feature to block updates when not in the maintenance window.

## Parameters

### Maintenance Windows

Maintenance windows are set on a per-device basis using [balena device tags](https://www.balena.io/docs/learn/manage/filters-tags/#device-tags). The accepted tag format is:
```
Tag Name: MAINTENANCE_WINDOW
Value: 17:00:00_23:00:00
```

Maintenance windows are always evaluated based on the container's system timezone, to ensure local times affected by things like DST are respected. The container timezone can be changed by using the `TIMEZONE` env var, and defaults to UTC.

Window values crossing midnight, e.g. `23:00:00_02:00:00`, are accepted. They operate "as expected" - in this case, "starting at 11PM and ending at 2AM".

### Environment Variables

balena internal env vars used:
```bash
BALENA_API_KEY
BALENA_DEVICE_UUID
```

Optional env vars respected:
```bash
TIMEZONE=America/Los_Angeles
LOCKFILE_LOCATION=/tmp/balena
CHECK_INTERVAL_SECONDS=60
LOG_LEVEL=panic|fatal|error|warn|info|debug|trace
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

### Tests

To run test suite (Dockerized):
```
make test
```

To run test suite locally:
```
make test-local
```
