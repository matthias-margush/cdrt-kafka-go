package crdt

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/pkg/errors"
	kafka "gopkg.in/Shopify/sarama.v1"
)

// todo: handle closing

// Message is a message payload
type Message []byte

// Propagator defines channels for syncing state between peers
type Propagator struct {
	topic    string
	Producer chan Message
	Consumer chan Message
}

// NewPropagator creates a propagator
func NewPropagator(brokerList []string, topic string) (*Propagator, error) {
	log.Printf("Creating propagator (brokerList: %v, topic: %v)\n", brokerList, topic)
	var config *kafka.Config = nil //kafka.NewConfig()
	producer, err := kafka.NewSyncProducer(brokerList, config)
	if err != nil {
		return nil, errors.Wrap(err, "creating producer")
	}

	consumer, err := kafka.NewConsumer(brokerList, config)
	if err != nil {
		return nil, errors.Wrap(err, "creating consumer")
	}

	partition, err := consumer.ConsumePartition(topic, 0, kafka.OffsetOldest)
	if err != nil {
		return nil, errors.Wrap(err, "creating partition consumer")
	}

	propagator := Propagator{
		topic:    topic,
		Producer: make(chan Message),
		Consumer: make(chan Message),
	}

	sigs := make(chan os.Signal)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	go propagator.propagate(producer, sigs)
	go propagator.sync(partition, sigs)

	return &propagator, nil
}

// propagate pushes updates to peers
func (propagator *Propagator) propagate(producer kafka.SyncProducer, sigs chan os.Signal) {
	for {
		select {
		case <-sigs:
			producer.Close()
			return
		case payload := <-propagator.Producer:
			message := &kafka.ProducerMessage{
				Topic: propagator.topic,
				Value: kafka.ByteEncoder(payload),
			}
			log.Printf("Sending message: %v\n", message)
			partition, offset, err := producer.SendMessage(message)
			if err != nil {
				log.Printf("Failed to send message: %v: %v\n", message, err)
			} else {
				log.Printf("Sent message (partition: %v, offset: %v): %v\n", partition, offset, message)
			}
		}
	}
}

// sync receives updates from peers
func (propagator *Propagator) sync(consumer kafka.PartitionConsumer, sigs chan os.Signal) {
	for {
		log.Printf("Consuming %v\n", consumer)
		select {
		case <-sigs:
			err := consumer.Close()
			if err != nil {
				log.Printf("Failed to close consumer: %s\n", err)
			}
			return
		case message := <-consumer.Messages():
			log.Printf("Message on %v: %v\n", consumer, message)
			propagator.Consumer <- message.Value
		}
	}
}
