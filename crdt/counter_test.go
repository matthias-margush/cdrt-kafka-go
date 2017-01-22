package crdt_test

import (
	"log"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/matthias-margush/crdt-kafka-go/crdt"
)

func brokerList() []string {
	brokerList := os.Getenv("BROKER_LIST")
	if brokerList == "" {
		brokerList = "localhost:9092"
	}
	log.Printf("brokerList: %v\n", brokerList)
	return strings.Split(brokerList, ",")
}

func TestNewCounter(t *testing.T) {
	propagationTopic := "Propagation-Test-" + uuid.New().String()
	propagator1, err := crdt.NewPropagator(brokerList(), propagationTopic)
	if err != nil {
		t.Fatalf("Failed to start propagator: %v\n", err)
	}

	propagator2, err := crdt.NewPropagator(brokerList(), propagationTopic)
	if err != nil {
		t.Fatalf("Failed to start propagator: %v\n", err)
	}

	c1 := crdt.NewCounter("c1", propagator1)
	c2 := crdt.NewCounter("c2", propagator2)

	c1.Inc() // 1
	c2.Inc() // 2
	c2.Inc() // 3
	c1.Dec() // 2
	c2.Dec() // 1

	time.Sleep(1 * time.Second) // let things sync and settle

	if c1.Val() != 1 {
		t.Errorf("Expected %v, actual %v", 1, c2.Val())
	}

	if c2.Val() != 1 {
		t.Errorf("Expected %v, actual %v", 1, c2.Val())
	}
}
