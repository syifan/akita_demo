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

- Go 1.22 or later

### Build

```bash
go build -o akita_demo .
```

### Run

```bash
./akita_demo
```

By default, the simulation will run for 20 seconds of simulated time. You can specify a different duration using the `-cycles` flag:

```bash
./akita_demo -cycles 10   # Run for 10 seconds
./akita_demo -cycles 50   # Run for 50 seconds
```

You can also control the consumer processing rate using the `-consume-rate` flag:

```bash
./akita_demo -consume-rate 0.5   # Consumers process 1 message every 0.5 seconds (faster)
./akita_demo -consume-rate 2.0   # Consumers process 1 message every 2 seconds (slower)
./akita_demo -cycles 30 -consume-rate 0.5   # Combine flags
```

The simulation will show:
- When the producer generates messages
- How the distributor routes messages to specific consumers
- When each consumer processes its messages

## Example Output

```
=== Starting Akita Demo Simulation ===
Simulation Duration: 20 cycles (seconds)
Producer: Randomly generates messages (30% chance per tick)
Distributor: Routes messages to correct consumer
Consumers: Process messages at rate (1 per 1.00 seconds)

[4.00] Producer: Generated message for Consumer1
[5.00] Distributor: Routed message to Consumer1
[6.00] Consumer Consumer1: Consumed message: Message at time 4.00
[7.00] Producer: Generated message for Consumer2
[8.00] Distributor: Routed message to Consumer2
...
```

## Command-Line Options

- `-cycles <number>`: Set the simulation duration in cycles (seconds). Default is 20.
  - Example: `./akita_demo -cycles 10`
- `-consume-rate <number>`: Set the consumer message processing rate (seconds between messages). Default is 1.0.
  - Example: `./akita_demo -consume-rate 0.5`
- `-h`: Display help message with all available options.

## Key Implementation Details

- Messages contain a `Destination` field specifying which consumer should receive them
- Producer generates traffic randomly (30% probability per tick)
- Distributor maintains separate output ports for each consumer
- Consumers enforce a configurable rate limit (default: 1 second between processing messages)
- All components are connected via Akita's DirectConnection
- The simulation uses ticking components that update every simulated second
