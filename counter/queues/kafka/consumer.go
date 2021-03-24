package kafka

import (
	"context"
	"io"
	"strings"
	"time"

	"github.com/segmentio/kafka-go"
)

type Consumer struct {
	reader *kafka.Reader
	stop   chan struct{}
}

func NewConsumer(brokers, topic string) *Consumer {
	return &Consumer{
		kafka.NewReader(kafka.ReaderConfig{
			Brokers: strings.Split(brokers, ","),
			Topic:   topic,
			GroupID: "counters",
		}),
		make(chan struct{}),
	}
}

func (c *Consumer) Listen(ctx context.Context, consumeFunc func(message []byte)) error {
	var err error
	listenCtx, cancel := context.WithTimeout(ctx, time.Second*5)
	defer cancel()

consume:
	for {
		select {
		case <-c.stop:
			break consume
		default:
			msg, err := c.reader.ReadMessage(listenCtx)
			if err != nil {
				// channel has been closed
				if err == io.EOF {
					err = nil
				}
				break consume
			}
			go consumeFunc(msg.Value)
		}
	}

	close(c.stop)

	return err
}

func (c *Consumer) Close() error {
	c.stop <- struct{}{}
	return c.reader.Close()
}
