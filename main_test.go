package main

import (
	"testing"

	"github.com/sarchlab/akita/v3/sim"
)

// TestDistributorReturnsFalseWhenNoMessages verifies that the distributor
// returns false when there are no messages to process
func TestDistributorReturnsFalseWhenNoMessages(t *testing.T) {
	engine := sim.NewSerialEngine()
	consumerNames := []string{"Consumer1"}
	distributor := NewDistributor("Distributor", engine, consumerNames)
	
	// Tick with no messages
	result := distributor.Tick(0)
	
	if result != false {
		t.Errorf("Expected Distributor.Tick() to return false when no messages, got %v", result)
	}
}

// TestConsumerReturnsFalseWhenNoMessages verifies that the consumer
// returns false when there are no messages to consume
func TestConsumerReturnsFalseWhenNoMessages(t *testing.T) {
	engine := sim.NewSerialEngine()
	consumer := NewConsumer("Consumer1", engine, 1.0)
	
	// Tick with no messages
	result := consumer.Tick(0)
	
	if result != false {
		t.Errorf("Expected Consumer.Tick() to return false when no messages, got %v", result)
	}
}

// TestConsumerReturnsFalseWhenRateLimiting verifies that the consumer
// returns false when rate limiting prevents consumption
func TestConsumerReturnsFalseWhenRateLimiting(t *testing.T) {
	engine := sim.NewSerialEngine()
	consumer := NewConsumer("Consumer1", engine, 1.0)
	
	// Simulate that a message was just consumed
	consumer.lastConsumed = 0
	
	// Create and send a message to the consumer
	msg := &DemoMessage{
		Content:     "Test message",
		Destination: "Consumer1",
	}
	msg.Meta().Src = nil
	msg.Meta().Dst = consumer.inputPort
	
	// Send the message
	consumer.inputPort.Recv(msg)
	
	// Try to tick before consume rate has passed (at time 0.5)
	result := consumer.Tick(0.5)
	
	if result != false {
		t.Errorf("Expected Consumer.Tick() to return false when rate limiting, got %v", result)
	}
}

// TestDistributorReturnsFalseWhenSendFails verifies that the distributor
// returns false when it fails to send a message (output port full)
func TestDistributorReturnsFalseWhenSendFails(t *testing.T) {
	engine := sim.NewSerialEngine()
	consumerNames := []string{"Consumer1"}
	distributor := NewDistributor("Distributor", engine, consumerNames)
	consumer := NewConsumer("Consumer1", engine, 1.0)
	
	// Register consumer port
	distributor.consumerPorts["Consumer1"] = consumer.inputPort
	
	// Connect distributor output to consumer input
	conn := sim.NewDirectConnection("TestConnection", engine, 1*sim.Hz)
	conn.PlugIn(distributor.outputPorts["Consumer1"], 1)
	conn.PlugIn(consumer.inputPort, 1)
	
	// Fill up the distributor's output port (capacity is 1)
	fillMsg := &DemoMessage{
		Content:     "Fill message",
		Destination: "Consumer1",
	}
	fillMsg.Meta().Src = distributor.outputPorts["Consumer1"]
	fillMsg.Meta().Dst = consumer.inputPort
	distributor.outputPorts["Consumer1"].Send(fillMsg)
	
	// Now send a message to distributor's input
	msg := &DemoMessage{
		Content:     "Test message",
		Destination: "Consumer1",
	}
	msg.Meta().Src = nil
	msg.Meta().Dst = distributor.inputPort
	distributor.inputPort.Recv(msg)
	
	// Try to tick - should fail because output port is full
	result := distributor.Tick(0)
	
	if result != false {
		t.Errorf("Expected Distributor.Tick() to return false when send fails, got %v", result)
	}
}

// TestDistributorContinuesTickingWhenMoreMessages verifies that the distributor
// returns true when it successfully processes a message and more are available
func TestDistributorContinuesTickingWhenMoreMessages(t *testing.T) {
	engine := sim.NewSerialEngine()
	consumerNames := []string{"Consumer1"}
	distributor := NewDistributor("Distributor", engine, consumerNames)
	consumer := NewConsumer("Consumer1", engine, 1.0)
	
	// Register consumer port
	distributor.consumerPorts["Consumer1"] = consumer.inputPort
	
	// Connect distributor output to consumer input
	conn := sim.NewDirectConnection("TestConnection", engine, 1*sim.Hz)
	conn.PlugIn(distributor.outputPorts["Consumer1"], 1)
	conn.PlugIn(consumer.inputPort, 1)
	
	// Send two messages to distributor's input
	msg1 := &DemoMessage{
		Content:     "Test message 1",
		Destination: "Consumer1",
	}
	msg1.Meta().Src = nil
	msg1.Meta().Dst = distributor.inputPort
	distributor.inputPort.Recv(msg1)
	
	msg2 := &DemoMessage{
		Content:     "Test message 2",
		Destination: "Consumer1",
	}
	msg2.Meta().Src = nil
	msg2.Meta().Dst = distributor.inputPort
	distributor.inputPort.Recv(msg2)
	
	// Tick - should return true because more messages are available
	result := distributor.Tick(0)
	
	if result != true {
		t.Errorf("Expected Distributor.Tick() to return true when more messages available, got %v", result)
	}
}
