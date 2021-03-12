package bus

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/segmentio/kafka-go"
	log "github.com/sirupsen/logrus"

	"github.com/yelsukov/otus-ha/news/domain/entities"
)

type Listener struct {
	ctx          context.Context
	brokers      []string
	topic        string
	busChan      chan *entities.Event
	consumersQty int
}

func NewBusListener(ctx context.Context, brokers string, topic string, ch chan *entities.Event, cq int) *Listener {
	return &Listener{ctx, strings.Split(brokers, ","), topic, ch, cq}
}

func (l *Listener) runEventsConsumer(group string, i int) {
	log.Infof("running consumer #%d for `%s` group...", i, group)
	r := kafka.NewReader(kafka.ReaderConfig{
		Brokers: l.brokers,
		Topic:   l.topic,
		GroupID: group,
		//Logger:  log.WithContext(l.ctx),
	})
	defer func() {
		log.Infof("closing consumer #%d for `%s`...\n", i, group)
		_ = r.Close()
	}()

	for {
		select {
		case <-l.ctx.Done():
			log.Info("stopping bus listener")
			return
		default:
			msg, err := r.ReadMessage(l.ctx)
			if err != nil {
				log.WithError(err).Error("could not read message. stopping bus listener")
				return
			}
			event, err := parseMessage(msg.Value)
			if err != nil {
				log.WithError(err).Error("could not parse message: ", string(msg.Value))
				continue
			}
			if event == nil {
				continue
			}
			l.busChan <- event
		}
	}
}

func (l *Listener) Listen() {
	log.Info("running events bus listener")
	for i := 0; i < l.consumersQty; i++ {
		go l.runEventsConsumer("event.listeners", i)
	}
	log.Info("bus listener has been started")
}

func parseMessage(msg []byte) (*entities.Event, error) {
	var event entities.Event
	err := json.Unmarshal(msg, &event)
	return &event, err
}
