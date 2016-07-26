package consumer

import (
	"errors"
	"fmt"

	"github.com/localhots/caller"
	"github.com/localhots/shezmu"
)

// Subscriber is the interface that is used by daemons to subscribe to messages.
type Subscriber interface {
	Subscribe(consumerName, topic string) Streamer
}

// Streamer is the interface that wraps message consumers. Error handling
// should be provided by the implementation. Feel free to panic.
type Streamer interface {
	Messages() <-chan []byte
	Close()
}

// Publisher is the interface that wraps message publishers. Error handling
// should be provided by the implementation. Feel free to panic.
type Publisher interface {
	Publish(topic string, msg []byte, meta interface{})
	Close()
}

// Consumer extends Shezmu's BaseDaemon with pub/sub features.
type Consumer struct {
	shezmu.BaseDaemon
	publisher  Publisher
	subscriber Subscriber
}

var (
	errMissingSubscriber = errors.New("subscriber is not set up")
	errMissingPublisher  = errors.New("publisher is not set up")
)

// Publish sends a message to the publisher.
func (c *Consumer) Publish(topic string, msg []byte, meta interface{}) {
	if c.publisher == nil {
		panic(errMissingPublisher)
	}

	c.publisher.Publish(topic, msg, meta)
}

// Subscribe subscriasdsdfsdgdfgdfsg sdgsdfg sdfgs dfgdfgdfg.
func (c *Consumer) Subscribe(topic string, fun interface{}) {
	name := fmt.Sprintf("subscription for topic %q", topic)
	c.SystemProcess(name, func() {
		if c.subscriber == nil {
			panic(errMissingSubscriber)
		}

		stream := c.subscriber.Subscribe(c.String(), topic)
		defer stream.Close()

		cf, err := caller.New(fun)
		if err != nil {
			panic(err)
		}

		for {
			select {
			case msg := <-stream.Messages():
				c.Process(func() { cf.Call(msg) })
			case <-c.ShutdownRequested():
				return
			}
		}
	})
}
