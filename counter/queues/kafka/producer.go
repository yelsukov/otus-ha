package kafka

import (
	"context"
	"github.com/segmentio/kafka-go"
)

type Producer struct {
	writer *kafka.Writer
}

func NewProducer(addr, topic string) *Producer {
	return &Producer{
		&kafka.Writer{
			Addr:     kafka.TCP(addr),
			Topic:    topic,
			Balancer: &kafka.LeastBytes{},
		},
	}
}

func (p *Producer) Publish(ctx context.Context, message []byte) error {
	return p.writer.WriteMessages(ctx, kafka.Message{Value: message})
}

func (p *Producer) Close() error {
	return p.writer.Close()
}
