#!/bin/bash

# Generate host keys if not present
ssh-keygen -A

# Start cron
service cron start

# Start SSH
/usr/sbin/sshd -D
