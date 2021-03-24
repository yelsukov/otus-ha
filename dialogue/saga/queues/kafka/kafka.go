package kafka

import (
	"context"
	"encoding/json"
	"io"
	"strings"

	"github.com/segmentio/kafka-go"
	log "github.com/sirupsen/logrus"

	"github.com/yelsukov/otus-ha/dialogue/domain/entities"
	"github.com/yelsukov/otus-ha/dialogue/saga/queues"
)

type BusQueue struct {
	writer *kafka.Writer
	reader *kafka.Reader
}

func NewBus(brokers, writerTopic, readerTopic string) *BusQueue {
	return &BusQueue{
		&kafka.Writer{
			Addr:     kafka.TCP(brokers),
			Topic:    writerTopic,
			Balancer: &kafka.LeastBytes{},
		},
		kafka.NewReader(kafka.ReaderConfig{
			Brokers: strings.Split(brokers, ","),
			Topic:   readerTopic,
			GroupID: "dialogues",
		}),
	}
}

func (k *BusQueue) Publish(ctx context.Context, saga *entities.Saga) error {
	payload, err := json.Marshal(queues.SagaCounterMessage{
		SagaId:  saga.Id,
		Command: saga.Command,
		ChatId:  saga.ChatId,
		UserId:  saga.UserId,
		Num:     saga.Num,
	})
	if err != nil {
		return err
	}

	log.Info("publishing event for saga " + saga.Id)
	return k.writer.WriteMessages(ctx, kafka.Message{Value: payload})
}

func (k *BusQueue) Listen(ctx context.Context, consumeFunc func(message queues.SagaInboundMessage)) {
	log.Info("running events listener")
	attempt := 0
consume:
	for {
		select {
		case <-ctx.Done():
			log.Info("stopping bus listener")
			return
		default:
			msg, err := k.reader.ReadMessage(ctx)
			if err != nil {
				if attempt > 5 || err == io.EOF {
					log.WithError(err).Error("could not read message. stopping bus listener")
					return
				}
				attempt++
				continue consume
			}

			attempt = 0
			var event queues.SagaInboundMessage
			if err = json.Unmarshal(msg.Value, &event); err != nil {
				log.WithError(err).Error("could not parse message: " + string(msg.Value))
				continue
			}
			if event.SagaId == "" {
				continue
			}
			consumeFunc(event)
		}
	}
}

func (k *BusQueue) Close() {
	if err := k.writer.Close(); err != nil {
		log.WithError(err).Error("cannot close producer conn")
	} else {
		log.Info("producer connection has been closed")
	}
	if err := k.reader.Close(); err != nil {
		log.WithError(err).Error("cannot close consumer conn")
	} else {
		log.Info("consumer connection has been closed")
	}
}
