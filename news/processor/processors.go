package processor

import (
	"context"

	log "github.com/sirupsen/logrus"

	"github.com/yelsukov/otus-ha/news/domain/entities"
	"github.com/yelsukov/otus-ha/news/domain/models"
)

// Custom processor for user login event
func (pm *ProcessorsManager) processLogin(ctx context.Context, ch chan *entities.Event, n int) {
	for {
		select {
		case <-ctx.Done():
			log.Infof("stopping processor #%d for `login` event", n)
			return
		case event := <-ch:
			// Store event at persistent storage
			if err := pm.storeEvents(event); err != nil {
				log.WithError(err).Error("fail on storing `login` event")
				continue
			}

			// Heat the events cache for user
			log.Debugf("Heating events for user#%d", event.ObjectId)
			go pm.heater.HeatEvents(event.ObjectId)
		}
	}
}

// Custom processor for user logout event
func (pm *ProcessorsManager) processLogout(ctx context.Context, ch chan *entities.Event, n int) {
	for {
		select {
		case <-ctx.Done():
			log.Infof("stopping processor #%d for `logout` event", n)
			return
		case event := <-ch:
			// Store event at persistent storage
			if err := pm.storeEvents(event); err != nil {
				log.WithError(err).Error("fail on storing `logout` event")
				continue
			}

			// Invalidate the events cache for user
			log.Debug("Invalidating cache on logout")
			go func() {
				if err := pm.cache.DeleteEvents(event.ObjectId); err != nil {
					log.WithError(err).Error("failed to invalidate cache")
				}
			}()
		}
	}
}

// Custom processor for follower add event
func (pm *ProcessorsManager) processFollow(ctx context.Context, ch chan *entities.Event, n int) {
	for {
		select {
		case <-ctx.Done():
			log.Infof("stopping processor #%d for `follow` event", n)
			return
		case event := <-ch:
			log.Debugf("storing new follower #%d for user #%d into db", event.SubjectId, event.ObjectId)
			err := pm.followerStorage.AddFollower(event.SubjectId, event.ObjectId)
			if err != nil {
				log.WithError(err).Error("fail on follower add")
				continue
			}

			// Store event at persistent storage
			if err := pm.storeEvents(event); err != nil {
				log.WithError(err).Error("fail on storing event")
				continue
			}

			// Add follower to cache
			go func() {
				if err = pm.cache.AddFollowers(event.SubjectId, event.ObjectId); err != nil {
					log.WithError(err).Error("fail on adding follower to cache")
				}
			}()
		}
	}
}

// Common events processor. Stores events at persistent db and at cache
func (pm *ProcessorsManager) storeEvents(event *entities.Event) error {
	// Read followers of event producer
	ff, err := pm.followerStorage.ReadOne(event.ObjectId)
	if err != nil {
		return err
	}
	if ff == nil || len(ff.List) == 0 {
		log.Debug("can not store event, producer have no followers")
		return nil
	}

	doc := models.Event{
		Name:       event.Name,
		Message:    event.Message,
		ProducerId: event.ObjectId,
		Followers:  ff.List,
		Timestamp:  event.Timestamp,
	}
	// Add event to persistent storage
	err = pm.eventStorage.InsertOne(&doc)
	if err != nil {
		return err
	}

	// Add event to cache of followers
	go pm.cache.AddEventToFollowers(&doc, ff.List...)

	pm.srvWriteCh <- &doc

	return nil
}
