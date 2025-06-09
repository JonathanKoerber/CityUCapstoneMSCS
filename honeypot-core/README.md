## Honeypot Core

The `honeypot-core` module powers an interactive, containerized cyber-physical honeypot designed to emulate Industrial Control Systems (ICS). This system creates realistic, multi-protocol targets that attract, delay, and analyze attackers in a controlled environment.

---

## Overview

This honeypot emulates an ICS network by combining:

- **FUXA** – A SCADA/HMI interface to visualize and interact with device data.
- **PLC Nodes** – Simulated devices that expose Modbus TCP interfaces and serve configurable device states.
- **Target System** – A "central" ICS host accessible over SSH and custom ports.
- **Deception & Analysis Services** – Optional LLM-backed tools to generate plausible system responses and analyze attacker behavior.

---

## Directory Structure

```bash
├── app
│   ├── data
│   │   ├── ssh
│   │   ├── ssh-context
│   │   └── walk_write_file.sh
│   ├── emulator
│   │   ├── chatSession.go
│   │   ├── emulator.go
│   │   ├── sshEmulator.go
│   │   └── vectorStore.go
│   ├── ics-node
│   │   ├── icsDevice.go
│   │   └── icsNode.go
│   ├── main.go
│   ├── plc-node
│   │   ├── Device-Config
│   │   ├── Dockerfile-Modbus-TCP
│   │   ├── go.mod
│   │   ├── go.sum
│   │   ├── main.go
│   │   ├── modbusServer
│   │   └── README.md
│   └── server
│       ├── server.go
│       └── sshServer.go
├── authorized_keys
├── Dockerfile-App
├── Dockerfile-Dev
├── go.mod
├── go.sum
├── honey
├── README.md
├── ssh_keys
│   ├── id_rsa
│   ├── id_rsa.pub
│   ├── mykey
│   └── mykey.pub
└── target
    ├── cron-jobs
    │   └── logger.sh
    ├── Dockerfile-Target
    ├── entrypoint.sh
    └── README.md
```
