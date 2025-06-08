#!/bin/bash

# Start SSH as root
service ssh start

# Make sure the cron directory is writable
mkdir -p /var/run/cron
chown root:root /var/run/cron

# Start cron as root
cron

# Keep container alive
tail -f /dev/null
