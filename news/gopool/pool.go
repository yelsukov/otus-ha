package gopool

import (
	"fmt"
	"time"

	log "github.com/sirupsen/logrus"
)

// ErrScheduleTimeout indicates that there no free goroutines during some period of time
var ErrScheduleTimeout = fmt.Errorf("schedule error: timed out")

// Pool contains logic of goroutine reuse.
type Pool struct {
	sem  chan struct{} // semaphore channel
	jobs chan func()   // channel for jobs
}

// NewPool creates new goroutine pool and jobs queue with given sizes.
// It also spawns given amount of goroutines
func NewPool(size, queue, spawn int) *Pool {
	p := &Pool{
		make(chan struct{}, size),
		make(chan func(), queue),
	}

	if spawn > size {
		spawn = size
	}

	for i := 0; i < spawn; i++ {
		p.sem <- struct{}{}
		go p.worker()
	}

	return p
}

// Schedule schedules job to be executed over pool's workers.
func (p *Pool) Schedule(job func()) {
	_ = p.schedule(job, nil)
}

// ScheduleTimeout schedules job to be executed over pool's workers.
// It returns ErrScheduleTimeout when no free workers met during given timeout.
func (p *Pool) ScheduleTimeout(timeout time.Duration, job func()) error {
	return p.schedule(job, time.After(timeout))
}

func (p *Pool) Close() {
	close(p.jobs)
	for len(p.sem) != 0 {
	}
	close(p.sem)
}

func (p *Pool) schedule(job func(), timeout <-chan time.Time) error {
	var err error
	select {
	case <-timeout:
		err = ErrScheduleTimeout
	case p.jobs <- job:
	case p.sem <- struct{}{}:
		log.Info("running new worker at pool")
		go p.worker()
		p.jobs <- job
	}
	return err
}

func (p *Pool) worker() {
	defer func() {
		<-p.sem
	}()

	for job := range p.jobs {
		job()
	}
}
