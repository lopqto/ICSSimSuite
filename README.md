# ICSSimSuite - ICS Simulation Suite

ICSSimSuite is an easy-to-use, open-source, and cross-platform simulation suite for Industrial Control Systems (ICS) with Modbus as the primary communication protocol. This suite allows users to simulate various ICS devices, such as HVAC systems, water pumps, batteries, and more, enabling testing and analysis of control systems in a safe and controlled environment.

## Table of Contents

- [Installation](#installation)
- [Usage](#usage)
- [Configuration](#configuration)
- [Simulated Devices](#simulated-devices)

## Installation

``` bash
go install github.com/lopqto/icssimsuite@latest
```

## Usage

``` bash
icssimsuite config.toml
```

## Configuration

Example configuration file can be found here: [config.toml](config.toml.example)

## Simulated Devices

### HVAC System (Unit ID: 1)

| Address | Description | Read/Write | Type | Function Code |
| --- | --- | --- | --- | --- |
| 1 | Fan State | R/W | bool | 0x01 (Coil) |
| 100 | Fan Speed | R/W | uint16 | 0x03 (Holding Register) |
| 100 | Temperature | R | float32 | 0x04 (Input Register) |
| 102 | Humidity | R | float32 | 0x04 (Input Register) |
| 104 | Room Temperature | R | float32 | 0x04 (Input Register) |
| 106 | Voltage | R | float32 | 0x04 (Input Register) |
| 108 | Current | R | float32 | 0x04 (Input Register) |
| 110 | Power | R | float32 | 0x04 (Input Register) |
| 200 | Uptime | R | uint32 | 0x04 (Input Register) |

### Pulse Counter (Unit ID: 2)

| Address | Description | Read/Write | Type | Function Code |
| --- | --- | --- | --- | --- |
| 0 | Pulse 1 State | R/W | bool | 0x01 (Coil) |
| 1 | Pulse 2 State | R/W | bool | 0x01 (Coil) |
| 2 | Pulse 3 State | R/W | bool | 0x01 (Coil) |
| 100 | Pulse 1 Count | R | uint32 | 0x04 (Input Register) | 
| 102 | Pulse 2 Count | R | uint32 | 0x04 (Input Register) |
| 104 | Pulse 3 Count | R | uint32 | 0x04 (Input Register) |

## Planned Devices 
- [ ] Water Pump
- [ ] Battery
- [ ] Solar Panel
- [ ] Wind Turbine

