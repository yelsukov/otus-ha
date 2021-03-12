package entities

import "time"

type Pool interface {
	Schedule(job func())
	ScheduleTimeout(timeout time.Duration, job func()) error
	Close()
}
