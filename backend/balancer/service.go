package balancer

import (
	"context"
	"errors"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"github.com/hashicorp/consul/api"
	log "github.com/sirupsen/logrus"
)

type Balancer struct {
	client      *api.Client
	serviceName string
	hosts       []string

	next uint32

	ticker *time.Ticker
	mu     sync.RWMutex
}

func New(consulDsn, serviceName string, hosts []string, interval time.Duration) (*Balancer, error) {
	consulCfg := api.DefaultConfig()
	consulCfg.Address = consulDsn
	client, err := api.NewClient(consulCfg)
	if err != nil {
		return nil, err
	}

	return &Balancer{
		client:      client,
		serviceName: serviceName,
		hosts:       hosts,
		ticker:      time.NewTicker(interval),
	}, nil
}

func (b *Balancer) RunHealthChecker(ctx context.Context) {
	var fails uint8
	for {
		select {
		case <-b.ticker.C:
			services, _, err := b.client.Health().Service(b.serviceName, "", false, nil)
			if err != nil {
				log.WithError(err).Error("fail on health check of " + b.serviceName)
				fails++
				if fails > 10 {
					log.WithError(err).Error("Consul is unavailable for long time. Health checker has been stopped for " + b.serviceName)
					return
				}
				continue
			}
			fails = 0

			if len(services) == len(b.hosts) {
				continue
			}

			hosts := make([]string, len(services), len(services))
			for i, s := range services {
				hosts[i] = s.Service.Address + ":" + strconv.Itoa(s.Service.Port)
			}
			b.mu.Lock()
			b.hosts = hosts
			b.mu.Unlock()
		case <-ctx.Done():
			log.Infof("health checker for %s has been stopped", b.serviceName)
			return
		}

	}
}

// GetAddr returns service address by round robin algorithm
func (b *Balancer) GetAddr() (string, error) {
	b.mu.Lock()
	defer b.mu.Unlock()

	if len(b.hosts) == 0 {
		return "", errors.New("have no available nodes for " + b.serviceName)
	}

	n := atomic.AddUint32(&b.next, 1)
	return b.hosts[(int(n)-1)%len(b.hosts)], nil
}
