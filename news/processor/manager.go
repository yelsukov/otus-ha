package processor

import (
	"context"

	log "github.com/sirupsen/logrus"

	"github.com/yelsukov/otus-ha/news/domain/entities"
	"github.com/yelsukov/otus-ha/news/domain/models"
	. "github.com/yelsukov/otus-ha/news/domain/storages"
)

type ProcessorsManager struct {
	ctx             context.Context
	BusChan         chan *entities.Event
	srvWriteCh      chan *models.Event
	cache           entities.Cache
	heater          entities.CacheHeater
	followerStorage FollowerStorage
	eventStorage    EventStorage
	procQty         int
}

func NewProcessorsManager(ctx context.Context, bus chan *entities.Event, write chan *models.Event, cache entities.Cache, heater entities.CacheHeater, fs FollowerStorage, es EventStorage, procQty int) *ProcessorsManager {
	return &ProcessorsManager{ctx, bus, write, cache, heater, fs, es, procQty}
}

func (pm *ProcessorsManager) StartProcessing() {
	log.Info("running processors manager")

	logInChan := make(chan *entities.Event, pm.procQty)
	logOutChan := make(chan *entities.Event, pm.procQty)
	followChan := make(chan *entities.Event, pm.procQty)
	storeChan := make(chan *entities.Event, pm.procQty)
	defer func() {
		log.Info("closing processor channels")
		close(logInChan)
		close(logOutChan)
		close(followChan)
		close(storeChan)
	}()

	for i := 0; i < pm.procQty; i++ {
		go pm.processFollow(pm.ctx, followChan, i)
		go pm.processLogin(pm.ctx, logInChan, i)
		go pm.processLogout(pm.ctx, logOutChan, i)
	}
	log.Info("processor manager started")

	for {
		select {
		case <-pm.ctx.Done():
			log.Info("shutting down the processor")
			return
		case event := <-pm.BusChan:
			switch event.Name {
			case "login":
				logInChan <- event
			case "logout":
				logOutChan <- event
			case "follow":
				followChan <- event
			default:
				log.Errorf("unknown event %v", event)
			}
		}
	}
}
