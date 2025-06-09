# ICS Honeypot Target Container (`target/`)

This directory contains the build context and `Dockerfile` for the **ICS target container**, which is part of the overall honeypot environment. 
This container simulates a vulnerable ICS host within the `ics-net` network and can be used for intrusion detection, adversary interaction, and behavior analysis.

## Purpose

The target container represents a **realistic ICS node** that an attacker might try to access or exploit.
It is networked with other Modbus TCP devices and SCADA software (like FUXA) to create a convincing, interactive honeypot environment.

## Tools Installed

This container has `net-tools` installed, which includes essential utilities to explore and debug network configurations:

- `ifconfig` – Show network interfaces and IP addresses
- `netstat` – View open ports, active connections, and listening services
- `route` – Show the routing table
- `arp` – Display ARP table entries
- `ping` – Test network connectivity to other containers
- `traceroute` – (optional) Trace the path packets take to a target

These tools help validate container networking, simulate real host activity, and provide visibility for monitoring.

## Example Network Commands

After starting the docker compose app, you can exec into it using ssh to the container. 

```bash
admin@0.0.0.0 -p 2222
```

This will the proxy pass you into this target. You can explore the network with these commands

```bash
ifconfig              # View IP addresses of all interfaces
netstat -tuln         # Show listening ports (TCP/UDP)
route -n              # Check routing table
arp -a                # See nearby devices' IP/MAC addresses
ping 172.38.1.10      # Test connectivity to influxdb container
```