package main

import (
	"context"
	"os"
	"os/signal"
	"runtime"
	"syscall"

	"github.com/siddontang/go-log/log"

	"github.com/yelsukov/otus-ha/replicator/conf"
	"github.com/yelsukov/otus-ha/replicator/sync"
)

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())

	sc := make(chan os.Signal, 1)
	signal.Notify(sc,
		os.Kill,
		os.Interrupt,
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT)

	cfg, err := conf.PopulateConfig()
	if err != nil {
		log.Fatalf("fail on config reading: %v", err)
	}

	r, err := sync.NewSync(cfg)
	if err != nil {
		log.Fatalf("fail on river start: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{}, 1)
	go func() {
		log.Info("running replica synchronizer")
		if err := r.Run(ctx); err != nil {
			log.Errorf("fail on synchronizer run: %v", err)
			cancel()
		}
		done <- struct{}{}
	}()

	select {
	case n := <-sc:
		log.Infof("receive signal %v, closing", n)
	case <-ctx.Done():
		log.Infof("context is done with %v, closing", ctx.Err())
	}

	log.Info("shutting down the replicator...")
	r.Close()
	<-done
	log.Info("replicator has been stopped")
}
