package main

import (
	"flag"
	"fmt"
	"log"
	"math/rand"
	"time"

	"github.com/sarchlab/akita/v3/sim"
)

// DemoMessage represents a message with a destination consumer
type DemoMessage struct {
	meta        sim.MsgMeta
	Content     string
	Destination string
}

// Meta returns the message metadata
func (m *DemoMessage) Meta() *sim.MsgMeta {
	return &m.meta
}

// Clone creates a copy of the message
func (m *DemoMessage) Clone() sim.Msg {
	clone := *m
	return &clone
}

// Producer generates messages randomly and sends to distributor
type Producer struct {
	*sim.TickingComponent
	outputPort sim.Port
	dstPort    sim.Port // Destination port (distributor's input)
	consumers  []string
	rand       *rand.Rand
	stopTime   sim.VTimeInSec
}

// NewProducer creates a new producer component
func NewProducer(name string, engine sim.Engine, consumers []string, stopTime sim.VTimeInSec) *Producer {
	p := &Producer{
		consumers: consumers,
		rand:      rand.New(rand.NewSource(time.Now().UnixNano())),
		stopTime:  stopTime,
	}
	p.TickingComponent = sim.NewTickingComponent(name, engine, 1*sim.Hz, p)
	p.outputPort = sim.NewLimitNumMsgPort(p, 1, name+".Out")
	return p
}

// Tick generates messages randomly
func (p *Producer) Tick(now sim.VTimeInSec) bool {
	// Stop generating after stopTime
	if now >= p.stopTime {
		return false
	}
	
	// Random generation: 30% chance to generate a message each tick
	if p.rand.Float64() < 0.3 {
		// Pick a random consumer as destination
		dest := p.consumers[p.rand.Intn(len(p.consumers))]
		
		msg := &DemoMessage{
			Content:     fmt.Sprintf("Message at time %.2f", now),
			Destination: dest,
		}
		msg.Meta().Src = p.outputPort
		msg.Meta().Dst = p.dstPort
		msg.Meta().SendTime = now
		
		err := p.outputPort.Send(msg)
		if err != nil {
			return false
		}
		fmt.Printf("[%.2f] Producer: Generated message for %s\n", now, dest)
	}
	return true
}

// Distributor routes messages to the correct consumer
type Distributor struct {
	*sim.TickingComponent
	inputPort      sim.Port
	outputPorts    map[string]sim.Port
	consumerPorts  map[string]sim.Port // Map from consumer name to their input ports
}

// NewDistributor creates a new distributor component
func NewDistributor(name string, engine sim.Engine, consumers []string) *Distributor {
	d := &Distributor{
		outputPorts:   make(map[string]sim.Port),
		consumerPorts: make(map[string]sim.Port),
	}
	d.TickingComponent = sim.NewTickingComponent(name, engine, 1*sim.Hz, d)
	d.inputPort = sim.NewLimitNumMsgPort(d, 10, name+".In")
	
	for _, consumer := range consumers {
		d.outputPorts[consumer] = sim.NewLimitNumMsgPort(d, 1, name+".Out."+consumer)
	}
	
	return d
}

// Tick processes messages from input and routes to output
func (d *Distributor) Tick(now sim.VTimeInSec) bool {
	msg := d.inputPort.Peek()
	if msg == nil {
		return true
	}
	
	demoMsg, ok := msg.(*DemoMessage)
	if !ok {
		d.inputPort.Retrieve(now)
		return true
	}
	
	outputPort, ok := d.outputPorts[demoMsg.Destination]
	if !ok {
		fmt.Printf("[%.2f] Distributor: Unknown destination %s\n", now, demoMsg.Destination)
		d.inputPort.Retrieve(now)
		return true
	}
	
	consumerPort, ok := d.consumerPorts[demoMsg.Destination]
	if !ok {
		fmt.Printf("[%.2f] Distributor: Consumer port not registered for %s\n", now, demoMsg.Destination)
		d.inputPort.Retrieve(now)
		return true
	}
	
	newMsg := &DemoMessage{
		Content:     demoMsg.Content,
		Destination: demoMsg.Destination,
	}
	newMsg.Meta().Src = outputPort
	newMsg.Meta().Dst = consumerPort
	newMsg.Meta().SendTime = now
	
	err := outputPort.Send(newMsg)
	if err == nil {
		d.inputPort.Retrieve(now)
		fmt.Printf("[%.2f] Distributor: Routed message to %s\n", now, demoMsg.Destination)
	}
	
	return true
}

// Consumer consumes messages at a fixed rate
type Consumer struct {
	*sim.TickingComponent
	inputPort     sim.Port
	name          string
	lastConsumed  sim.VTimeInSec
	consumeRate   sim.VTimeInSec // Time between consuming messages
}

// NewConsumer creates a new consumer component
func NewConsumer(name string, engine sim.Engine, consumeRate sim.VTimeInSec) *Consumer {
	c := &Consumer{
		name:          name,
		consumeRate:   consumeRate,
		lastConsumed:  -1000, // Start with a large negative value
	}
	c.TickingComponent = sim.NewTickingComponent(name, engine, 1*sim.Hz, c)
	c.inputPort = sim.NewLimitNumMsgPort(c, 10, name+".In")
	return c
}

// Tick processes messages at a fixed rate
func (c *Consumer) Tick(now sim.VTimeInSec) bool {
	// Check if enough time has passed since last consumption
	if now-c.lastConsumed < c.consumeRate {
		return true
	}
	
	msg := c.inputPort.Peek()
	if msg == nil {
		return true
	}
	
	demoMsg, ok := msg.(*DemoMessage)
	if !ok {
		c.inputPort.Retrieve(now)
		return true
	}
	
	c.inputPort.Retrieve(now)
	c.lastConsumed = now
	fmt.Printf("[%.2f] Consumer %s: Consumed message: %s\n", now, c.name, demoMsg.Content)
	
	return true
}

func main() {
	// Parse command-line flags
	cycles := flag.Int("cycles", 20, "Number of simulation cycles (seconds) to run")
	flag.Parse()
	
	// Validate cycles value
	if *cycles <= 0 {
		log.Fatal("Error: cycles must be a positive number")
	}
	
	// Create simulation engine
	engine := sim.NewSerialEngine()
	
	// Define consumers
	consumerNames := []string{"Consumer1", "Consumer2", "Consumer3"}
	
	// Create components with configurable stop time
	producer := NewProducer("Producer", engine, consumerNames, sim.VTimeInSec(*cycles))
	distributor := NewDistributor("Distributor", engine, consumerNames)
	
	// Set producer's destination to distributor's input port
	producer.dstPort = distributor.inputPort
	
	// Create consumers with fixed consumption rate (1 message per second)
	consumers := make([]*Consumer, len(consumerNames))
	for i, name := range consumerNames {
		consumers[i] = NewConsumer(name, engine, 1.0) // 1 second between messages
	}
	
	// Connect producer to distributor
	conn := sim.NewDirectConnection("ProducerToDistributor", engine, 1*sim.Hz)
	conn.PlugIn(producer.outputPort, 1)
	conn.PlugIn(distributor.inputPort, 1)
	
	// Connect distributor to consumers
	for i, consumer := range consumers {
		// Register consumer input port with distributor
		distributor.consumerPorts[consumerNames[i]] = consumer.inputPort
		
		conn := sim.NewDirectConnection(
			fmt.Sprintf("DistributorTo%s", consumerNames[i]),
			engine,
			1*sim.Hz,
		)
		conn.PlugIn(distributor.outputPorts[consumerNames[i]], 1)
		conn.PlugIn(consumer.inputPort, 1)
	}
	
	// Kick off the ticking components
	producer.TickNow(0)
	distributor.TickNow(0)
	for _, consumer := range consumers {
		consumer.TickNow(0)
	}
	
	// Run simulation
	fmt.Println("=== Starting Akita Demo Simulation ===")
	fmt.Printf("Simulation Duration: %d cycles (seconds)\n", *cycles)
	fmt.Println("Producer: Randomly generates messages (30% chance per tick)")
	fmt.Println("Distributor: Routes messages to correct consumer")
	fmt.Println("Consumers: Process messages at fixed rate (1 per second)")
	fmt.Println()
	
	err := engine.Run()
	if err != nil {
		log.Fatal(err)
	}
	
	fmt.Println("\n=== Simulation Complete ===")
}
