package kafka

import (
	"context"
	"io"
	"strings"

	"github.com/segmentio/kafka-go"
)

type Consumer struct {
	reader *kafka.Reader
}

func NewConsumer(brokers, topic string) *Consumer {
	return &Consumer{
		kafka.NewReader(kafka.ReaderConfig{
			Brokers: strings.Split(brokers, ","),
			Topic:   topic,
			GroupID: "counters",
		}),
	}
}

func (c *Consumer) Listen(ctx context.Context, consumeFunc func(message []byte)) error {
	var err error
	var attempt = 0
	for {
		msg, err := c.reader.ReadMessage(ctx)
		if err != nil {
			// connect has been closed
			if err == io.EOF {
				err = nil
				break
			}
			if attempt > 5 {
				break
			}
			attempt++
			continue
		}
		attempt = 0
		go consumeFunc(msg.Value)
	}

	return err
}

func (c *Consumer) Close() error {
	go func() {
		_ = c.reader.Close()
	}()
	return nil
}
