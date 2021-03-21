package zabbix

import (
	"context"
	"fmt"
	"github.com/yelsukov/otus-ha/dialogue/config"
	"strconv"
	"time"

	. "github.com/blacked/go-zabbix"
	"github.com/mackerelio/go-osstat/cpu"
	"github.com/mackerelio/go-osstat/memory"
	log "github.com/sirupsen/logrus"
)

const MaxFailsNum = 5
const ObservationInterval = time.Second * 5

func ObserveMetrics(ctx context.Context, cfg *config.Config) {
	sender := NewSender(cfg.ZabbixHost, cfg.ZabbixPort)
	ticker := time.NewTicker(ObservationInterval)
	defer ticker.Stop()
	attempts := 0
	for {
		select {
		case <-ticker.C:
			pkg, err := metrics(cfg.ZabbixName, cfg.ServiceId)
			if err != nil {
				log.WithError(err).Error("Failed to collect metrics")
				continue
			}

			if _, err = sender.Send(pkg); err != nil {
				attempts++
				if attempts > MaxFailsNum {
					log.WithError(err).Error("zabbix observer stopped due to a lot of errors")
					return
				}
				log.WithError(err).Error("failed to send metrics to zabbix")
			} else {
				attempts = 0
			}
		case <-ctx.Done():
			log.Info("Zabbix observer has been stopped")
			return
		}
	}
}

func metrics(host, id string) (*Packet, error) {
	mem, err := memory.Get()
	if err != nil {
		return nil, err
	}

	before, err := cpu.Get()
	if err != nil {
		return nil, err
	}
	time.Sleep(time.Second * 1)
	after, err := cpu.Get()
	if err != nil {
		return nil, err
	}
	total := float64(after.Total - before.Total)

	var metrics = []*Metric{
		NewMetric(host, id+"-mem-used", strconv.FormatUint(mem.Used, 10),
			time.Now().Unix()),
		NewMetric(host, id+"-mem-cache", strconv.FormatUint(mem.Cached, 10),
			time.Now().Unix()),
		NewMetric(host, id+"-mem-free", strconv.FormatUint(mem.Free, 10),
			time.Now().Unix()),
		NewMetric(host, id+"-cpu-used", fmt.Sprintf("%f", float64(after.User-before.User)/total*100),
			time.Now().Unix()),
		NewMetric(host, id+"-cpu-system", fmt.Sprintf("%f", float64(after.System-before.System)/total*100),
			time.Now().Unix()),
		NewMetric(host, id+"-cpu-idle", fmt.Sprintf("%f", float64(after.Idle-before.Idle)/total*100),
			time.Now().Unix()),
	}

	return NewPacket(metrics), nil
}
