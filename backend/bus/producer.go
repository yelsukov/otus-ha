package bus

import (
	"context"
	"encoding/json"
	"time"

	"github.com/segmentio/kafka-go"
	log "github.com/sirupsen/logrus"
)

type Producer struct {
	ctx    context.Context
	writer *kafka.Writer
}

func NewProducer(ctx context.Context, dsn, topic string) *Producer {
	return &Producer{
		ctx,
		&kafka.Writer{
			Addr:     kafka.TCP(dsn),
			Topic:    topic,
			Balancer: &kafka.LeastBytes{},
		},
	}
}

func (p *Producer) Close() {
	log.Info("closing producer")
	if err := p.writer.Close(); err == nil {
		log.WithError(err).Error("cannot close kafka conn")
	} else {
		log.Info("producer connection has been closed")
	}
}

type eventMessage struct {
	Name      string `json:"name"`
	Message   string `json:"msg"`
	ObjectId  int    `json:"oid"`
	SubjectId int    `json:"sid"`
	Timestamp int64  `json:"ts"`
}

func (p *Producer) WriteEvent(name, message string, producerId, subjectId int) {
	msg, err := json.Marshal(&eventMessage{
		name,
		message,
		producerId,
		subjectId,
		time.Now().Unix(),
	})
	if err != nil {
		log.WithError(err).Error("failed to write event into the bus")
		return
	}
	if err = p.writer.WriteMessages(p.ctx, kafka.Message{Value: msg}); err != nil {
		log.WithError(err).Error("failed to write event into the bus")
	}
}
