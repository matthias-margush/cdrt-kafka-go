package crdt

import (
	"encoding/json"
	"log"
)

// Counter is an op-based CmRDT counter (section 3.1.1)
type Counter struct {
	name       string
	value      int
	propagator *Propagator
}

const (
	add int = iota
	sub     = iota
)

type op struct {
	Name interface{}
	Amt  int
}

// NewCounter creates a new counter
func NewCounter(name string, propagator *Propagator) *Counter {
	counter := Counter{
		name:       name,
		propagator: propagator,
	}

	go counter.sync()
	return &counter
}

// Inc increments the counter by 1
func (c *Counter) Inc() {
	c.propagate(op{c.name, 1})
	c.value++
}

// Dec decrements the counter by 1
func (c *Counter) Dec() {
	c.propagate(op{c.name, -1})
	c.value--
}

// Val gets the current value of the counter
func (c *Counter) Val() int {
	return c.value
}

// propagate sends updates to peers
func (c *Counter) propagate(op op) {
	payload, err := json.Marshal(op)
	if err != nil {
		log.Fatalf("Failed to marshal message: %v\n", err)
	}
	c.propagator.Producer <- payload
}

// sync listens for updates from peers
func (c *Counter) sync() {
	log.Printf("Starting syncing: %v", c)
	for payload := range c.propagator.Consumer {
		op := op{}
		err := json.Unmarshal(payload, &op)
		if err != nil {
			log.Fatalf("Failed to unmarshal message: %v\n", err)
		}

		if op.Name != c.name {
			c.value += op.Amt
			log.Printf("Synced %v with op: %v\n", c, op)
		} else {
			log.Printf("Skipped %v with op: %v\n", c, op)
		}
	}
}
