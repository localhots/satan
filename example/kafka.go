package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"github.com/Shopify/sarama"
	"github.com/localhots/satan"
)

// KafkaConsumerState contains data that is required to create a Kafka consumer.
type KafkaConsumerState struct {
	Partition int32 `json:"partition"`
	Offset    int64 `json:"offset"`
}

// KafkaStream is an implementation of satan.Stremer for Kafka messaging queue.
type KafkaStream struct {
	messages chan []byte
	shutdown chan struct{}
}

const (
	consumerStateFile = "tmp/consumers.json"
)

var (
	kafkaClient   sarama.Client
	kafkaConsumer sarama.Consumer
	consumers     = map[string]map[string]KafkaConsumerState{}
)

func initKafka(brokers []string) {
	conf := sarama.NewConfig()
	conf.ClientID = "Satan Example"

	var err error
	if kafkaClient, err = sarama.NewClient(brokers, conf); err != nil {
		panic(err)
	}
	if kafkaConsumer, err = sarama.NewConsumerFromClient(kafkaClient); err != nil {
		panic(err)
	}
}

func shutdownKafka() {
	if err := kafkaConsumer.Close(); err != nil {
		panic(err)
	}
	if err := kafkaClient.Close(); err != nil {
		panic(err)
	}
}

func makeStream(consumer, topic string) satan.Streamer {
	c, ok := consumers[consumer]
	if !ok {
		panic(fmt.Errorf("Consumer %q has no config", consumer))
	}
	t, ok := c[topic]
	if !ok {
		panic(fmt.Errorf("Consumer %q has no config for topic %q", consumer, topic))
	}

	pc, err := kafkaConsumer.ConsumePartition(topic, t.Partition, t.Offset)
	if err != nil {
		panic(err)
	}

	stream := &KafkaStream{
		messages: make(chan []byte),
		shutdown: make(chan struct{}),
	}
	go func() {
		for {
			select {
			case msg := <-pc.Messages():
				stream.messages <- msg.Value
				t.Offset = msg.Offset
			case err := <-pc.Errors():
				log.Println("Kafka error:", err.Error())
			case <-stream.shutdown:
				return
			}
		}
	}()

	return stream
}

func loadConsumerConfig() {
	if b, err := ioutil.ReadFile(consumerStateFile); err != nil {
		fmt.Println(`Kafka consumers state file was not found at ` + consumerStateFile + `
Please create one in order to proceed with this example.
Config file contents should look like this:
{
    "ConsumerName": {
        "TopicName": {
            "partition": 0,
            "offset": 12345
        }
    }
}`)
		os.Exit(1)
	} else {
		if err = json.Unmarshal(b, &consumers); err != nil {
			panic(err)
		}
	}
}

// Messages returns a channel that stream messages.
func (s *KafkaStream) Messages() <-chan []byte {
	return s.messages
}

// Close stops Kafka partition consumer.
func (s *KafkaStream) Close() {
	close(s.shutdown)
}
