package storages

import (
	"context"
	"encoding/json"

	log "github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/yelsukov/otus-ha/news/domain/entities"
	"github.com/yelsukov/otus-ha/news/domain/models"
)

type EventStorage struct {
	ctx   context.Context
	col   *mongo.Collection
	cache entities.Cache
}

func NewEventStorage(ctx context.Context, db *mongo.Database, cache entities.Cache) *EventStorage {
	return &EventStorage{ctx, db.Collection("events"), cache}
}

func (c *EventStorage) InsertOne(event *models.Event) error {
	_, err := c.col.InsertOne(c.ctx, event)
	return err
}

func (c *EventStorage) ReadMany(uid int) ([]*models.Event, error) {
	var events []*models.Event

	ee, err := c.cache.ReadEvents(uid)
	if len(ee) != 0 {
		log.Debugf("got %d events from Cache for user #%d", len(ee), uid)
		events = make([]*models.Event, len(ee), len(ee))
		for i, e := range ee {
			var event models.Event
			if err = json.Unmarshal([]byte(e), &event); err != nil {
				log.WithError(err).Error("failed to parse event from cache")
				events = nil
				break
			}
			events[i] = &event
		}
	}

	if len(events) == 0 {
		opts := options.Find().SetSort(bson.D{{"dt", -1}}).SetLimit(1000)
		cursor, err := c.col.Find(c.ctx, bson.D{{"flr", uid}}, opts)
		if err != nil {
			return nil, err
		}
		if err = cursor.All(c.ctx, &events); err != nil {
			return nil, err
		}
		log.Debugf("got %d events from DB for user #%d", len(events), uid)
		if len(events) != 0 {
			c.cache.AddEvents(uid, events...)
		}
	}

	return events, nil
}
