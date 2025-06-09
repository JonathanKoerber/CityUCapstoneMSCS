# PLC Node Module (`plc-node/`)

This module contains the code and configuration necessary to simulate Programmable Logic Controllers (PLCs) in a Modbus TCP-based industrial control system (ICS) honeypot. Each container built from this module acts as a standalone device with unique behavior defined in a JSON config.

##  Purpose

The goal of this module is to emulate industrial field devices—like pumps, valves, and sensors—over Modbus TCP. These nodes are deployable as individual containers and can be discovered and interacted with by SCADA tools like FUXA or adversaries probing the network.

##  Directory Structure

```bash
.
├── Device-Config
│   ├── pump_unit_1.json
│   ├── pump_unit_2.json
│   └── pump_unit_3.json
├── Dockerfile-Modbus-TCP
├── go.mod
├── go.sum
├── main.go
├── modbusServer
│   ├── Device.go
│   └── modbusServer.go
└── README.md
```

## How It Works


run the in a container:

`sudo docker build -f ./honeypot-core/app/plc-node/Dockerfile-Modbus-TCP -t modbus-node:latest ./honeypot-core/app/plc-node`

run images 

`sudo docker run -d --name pumpTest --network honeynet --ip 172.18.0.15 modbus-node:latest`

Test ports

`mbpoll -m tcp -a 1 -r 0 -p 1502 -c 10 172.18.0.15`

Each device that you want to get a response from will need an entry in the project docker compose file in the project root
dir. See the example below. You need to make sure that the device has a unique ip address that in the subnet that is allocated 
to the container.

```yaml
device01:
  container_name: pump01
  build:
    context: ./honeypot-core/app/plc-node
    dockerfile: Dockerfile-Modbus-TCP
  environment:
    CONTEXT_PATH: "/app/Device-Config/pump_unit_1.json"
  expose:
    - "502"
  volumes:
    - ./honeypot-core/app/plc-node/Device-Config:/app/Device-Config
  networks:
    ics-net:
      ipv4_address: 172.38.1.20
```

Each device that also needs an entry in the Device-Config dir. The file path need to be added to the docker compose declaration 
as well. The deviceId need to be unique you will use this in Fuxa to identify the device that is being targeted.  

```
[
    {
        "deviceId": "101",
        "deviceName": "Temp",
        "lowerBound": "0",
        "lowerWarn": "32",
        "upperBound": "250",
        "upperWarn": "230",
        "target": "200",
        "metaData": {
            "context": "at the begining",
            "processNeighbors": ["102"]
        }
    }
]
```