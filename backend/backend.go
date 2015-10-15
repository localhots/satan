package backend

// Backend is the interface that should be implemented by Satan backends.
type Backend interface {
	NewConsumer(topic string, partition int32, offset int64) (c Consumer, err error)
	Publish(topic string, msg []byte) error
	Close() error
}

// Consumer is the interface that should be implemented by backend consumer.
type Consumer interface {
	NextMessage() (msg []byte, err error)
}
