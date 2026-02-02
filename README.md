# akita_demo

A demonstration program using the [Akita](https://github.com/sarchlab/akita) discrete event simulation framework.

## Overview

This demo implements a message-passing system with three types of components:

- **Producer**: Randomly generates messages and sends them to the distributor (30% chance per simulation tick)
- **Distributor**: Routes messages to the correct consumer based on the destination field in each message
- **Consumer** (3 instances): Consumes messages at a fixed rate (1 message per second)

## Architecture

```
Producer (random traffic)
    |
    | DemoMessage (with destination field)
    v
Distributor (routes by destination)
    |
    +-- Consumer1
    +-- Consumer2
    +-- Consumer3
       (fixed consumption rate: 1 msg/sec)
```

## Building and Running

### Prerequisites

- Go 1.16 or later

### Build

```bash
go build -o akita_demo .
```

### Run

```bash
./akita_demo
```

The simulation will run for 20 seconds of simulated time, showing:
- When the producer generates messages
- How the distributor routes messages to specific consumers
- When each consumer processes its messages

## Example Output

```
=== Starting Akita Demo Simulation ===
Producer: Randomly generates messages (30% chance per tick)
Distributor: Routes messages to correct consumer
Consumers: Process messages at fixed rate (1 per second)

[4.00] Producer: Generated message for Consumer1
[5.00] Distributor: Routed message to Consumer1
[6.00] Consumer Consumer1: Consumed message: Message at time 4.00
[7.00] Producer: Generated message for Consumer2
[8.00] Distributor: Routed message to Consumer2
...
```

## Key Implementation Details

- Messages contain a `Destination` field specifying which consumer should receive them
- Producer generates traffic randomly (30% probability per tick)
- Distributor maintains separate output ports for each consumer
- Consumers enforce a fixed rate limit (1 second between processing messages)
- All components are connected via Akita's DirectConnection
- The simulation uses ticking components that update every simulated second
