package kafka

import (
	"context"
	"encoding/json"
	"github.com/segmentio/kafka-go"
	log "github.com/sirupsen/logrus"
	"github.com/yelsukov/otus-ha/dialogue/domain/entities"
	"github.com/yelsukov/otus-ha/dialogue/saga/queues"
	"strings"
)

type KafkaQueue struct {
	brokers []string
	writer  *kafka.Writer
	topic   string
}

func NewListener(brokers, topic string) *KafkaQueue {
	return &KafkaQueue{
		strings.Split(brokers, ","),
		&kafka.Writer{
			Addr:     kafka.TCP(brokers),
			Topic:    topic,
			Balancer: &kafka.LeastBytes{},
		},
		topic,
	}
}

func (k *KafkaQueue) Publish(ctx context.Context, saga *entities.Saga) error {
	payload, err := json.Marshal(queues.SagaCounterMessage{
		SagaId:  saga.Id,
		Command: saga.CounterTrx.Command,
		ChatId:  saga.CounterTrx.ChatId,
		UserId:  saga.CounterTrx.UserId,
		Num:     saga.CounterTrx.Num,
	})
	if err != nil {
		log.WithError(err).Error("failed to write event into the bus")
		return err
	}

	return k.writer.WriteMessages(ctx, kafka.Message{Value: payload})
}

func (k *KafkaQueue) Listen(ctx context.Context, consumeFunc func(message queues.SagaInboundMessage)) {
	log.Info("running events listener")
	r := kafka.NewReader(kafka.ReaderConfig{
		Brokers: k.brokers,
		Topic:   k.topic,
		GroupID: "dlgCons",
	})
	defer func() {
		log.Info("closing consumer for `dlgCons` group...")
		_ = r.Close()
	}()

	for {
		select {
		case <-ctx.Done():
			log.Info("stopping bus listener")
			return
		default:
			msg, err := r.ReadMessage(ctx)
			if err != nil {
				log.WithError(err).Error("could not read message. stopping bus listener")
				return
			}
			var event queues.SagaInboundMessage
			if err = json.Unmarshal(msg.Value, &event); err != nil {
				log.WithError(err).Error("could not parse message: ", string(msg.Value))
				continue
			}
			if event.SagaId == "" {
				continue
			}
			consumeFunc(event)
		}
	}
}

func (k *KafkaQueue) Close() {
	if err := k.writer.Close(); err != nil {
		log.WithError(err).Error("cannot close kafka conn")
	} else {
		log.Info("producer connection has been closed")
	}
}
