#!/bin/bash

# Set local timezone using env var
echo "Using timezone: $TIMEZONE"
ln -fs /usr/share/zoneinfo/$TIMEZONE /etc/localtime
dpkg-reconfigure -f noninteractive tzdata

# Run app
balena-maintenance-windows
