package app

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	log "github.com/sirupsen/logrus"
)

func cacheKey(cid string, uid int) string {
	return fmt.Sprintf("cnt:%s:%d", cid, uid)
}

func (a *App) consume(message []byte) {
	var saga SagaInboundMessage
	if err := json.Unmarshal(message, &saga); err != nil {
		log.WithError(err).Error("failed to parse inbound message")
		return
	}

	outbound := SagaOutboundMessage{
		SagaId: saga.SagaId,
		Action: saga.Command,
		Status: statusSuccess,
	}
	key := cacheKey(saga.ChatId, saga.UserId)
	switch saga.Command {
	case cmdIncr:
		if err := a.increment(key, saga.Num); err != nil {
			log.WithError(err).Error("failed to increment counter " + key)
			outbound.Status = statusAbort
		}
	case cmdDecr:
		if err := a.decrement(key, saga.Num); err != nil {
			log.WithError(err).Error("failed to decrement counter " + key)
			outbound.Status = statusAbort
		}
	}
}

func (a *App) increment(key string, num uint) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*1)
	defer cancel()

	_, err := a.storage.Incr(ctx, key, int64(num))

	return err
}

func (a *App) decrement(key string, num uint) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*1)
	defer cancel()

	_, err := a.storage.Incr(ctx, key, int64(num))

	return err
}
