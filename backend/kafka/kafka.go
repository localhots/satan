package kafka

import (
	"errors"
	"fmt"

	"github.com/Shopify/sarama"
	"github.com/localhots/satan/backend"
)

// Client is the Kafka adapter client for Satan.
type Client struct {
	client     sarama.Client
	producer   sarama.SyncProducer
	consumer   sarama.Consumer
	pconsumers map[string]*Consumer
	shutdown   chan struct{}
}

// Consumer is a wrapper for Kafka partition consumer.
type Consumer struct {
	Topic     string
	Partition int32
	Offset    int64

	consumer sarama.PartitionConsumer
	shutdown chan struct{}
}

var (
	// ErrConsumerNotInitialized is returned when NewConsumer is called before
	// initializing Kafka consumer.
	ErrConsumerNotInitialized = errors.New("consumer is not initialized")
	// ErrProducerNotInitialized is returned when Publish is called before
	// initializing Kafka producer.
	ErrProducerNotInitialized = errors.New("producer is not initialized")
)

// New creates a new instance of Client.
func New(id string, brokers []string) (c *Client, err error) {
	conf := sarama.NewConfig()
	conf.ClientID = id
	conf.Consumer.Return.Errors = true

	client, err := sarama.NewClient(brokers, conf)
	if err != nil {
		return nil, err
	}

	return &Client{
		client:     client,
		pconsumers: make(map[string]*Consumer),
		shutdown:   make(chan struct{}),
	}, nil
}

// InitializeProducer initializes Kafka producer.
func (c *Client) InitializeProducer() error {
	if c.producer == nil {
		var err error
		if c.producer, err = sarama.NewSyncProducerFromClient(c.client); err != nil {
			return err
		}
	}

	return nil
}

// InitializeConsumer initializes Kafka consumer.
func (c *Client) InitializeConsumer() error {
	if c.consumer == nil {
		var err error
		if c.consumer, err = sarama.NewConsumerFromClient(c.client); err != nil {
			return err
		}
	}

	return nil
}

// NewConsumer creates a new partition consumer.
func (c *Client) NewConsumer(topic string, partition int32, offset int64) (cb backend.Consumer, err error) {
	if c.consumer == nil {
		return nil, ErrConsumerNotInitialized
	}

	pcons, err := c.consumer.ConsumePartition(topic, partition, offset)
	if err != nil {
		return nil, err
	}

	cons := &Consumer{
		Topic:     topic,
		Partition: partition,
		Offset:    offset,
		consumer:  pcons,
		shutdown:  make(chan struct{}),
	}

	name := fmt.Sprintf("%s-%d-#%d", topic, partition, len(c.pconsumers)+1)
	c.pconsumers[name] = cons

	return cons, nil
}

// Publish sends a message to producer.
func (c *Client) Publish(topic string, body []byte) error {
	if c.producer == nil {
		return ErrProducerNotInitialized
	}

	msg := &sarama.ProducerMessage{
		Topic: topic,
		Value: sarama.ByteEncoder(body),
	}
	_, _, err := c.producer.SendMessage(msg)

	return err
}

// Close shuts down Kafka producer and consumers.
func (c *Client) Close() error {
	for _, pc := range c.pconsumers {
		if err := pc.Close(); err != nil {
			return err
		}
	}
	if err := c.consumer.Close(); err != nil {
		return err
	}
	if err := c.producer.Close(); err != nil {
		return err
	}

	return nil
}

// NextMessage returns a pair of next message and next error. If both are nil,
// there are no messages left.
func (c *Consumer) NextMessage() (msg []byte, err error) {
	select {
	case err := <-c.consumer.Errors():
		return nil, err.Err
	case msg := <-c.consumer.Messages():
		c.Offset = msg.Offset
		return msg.Value, nil
	case <-c.shutdown:
		return nil, nil
	}
}

// Close shuts down partition consumer.
func (c *Consumer) Close() error {
	return c.consumer.Close()
}
