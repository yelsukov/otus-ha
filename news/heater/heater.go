package heater

import (
	log "github.com/sirupsen/logrus"
	"github.com/yelsukov/otus-ha/news/domain/entities"
	"github.com/yelsukov/otus-ha/news/domain/storages"
)

type CacheHeater struct {
	followers storages.FollowerStorage
	events    storages.EventStorage
	cache     entities.Cache
}

func NewCacheHeater(fs storages.FollowerStorage, es storages.EventStorage, cache entities.Cache) *CacheHeater {
	return &CacheHeater{fs, es, cache}
}

func (ch *CacheHeater) HeatFollowers() {
	log.Info("heating followers cache")
	ff, err := ch.followers.ReadMany()
	if err != nil {
		log.WithError(err).Error("fail to read followers on heating cache")
		return
	}
	for _, f := range ff {
		err = ch.cache.AddFollowers(f.Id, f.List...)
		if err != nil {
			log.WithError(err).Errorf("cannot heat cache for %+v", f)
		}
	}
	log.Info("cache heater has been run")
}

func (ch *CacheHeater) HeatEvents(uid int) {
	err := ch.cache.DeleteEvents(uid)
	if err != nil {
		log.WithError(err).Error("fail to delete events on heating cache")
	}
	// Read will create cache with events automatically
	events, err := ch.events.ReadMany(uid)
	if err != nil {
		log.WithError(err).Error("fail to read events on heating cache")
		return
	}

	log.Debugf("%d events where been added to cache", len(events))
}
