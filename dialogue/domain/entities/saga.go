package entities

import (
	"context"
)

type DialogTrx struct {
	MessagesIds []string `json:"ids"`
}

type CounterTrx struct {
	Command string `json:"cmd"`
	ChatId  string `json:"cid"`
	UserId  int    `json:"uid"`
	Num     uint   `json:"num"`
}

type Saga struct {
	Id         string `json:"-"`
	DialogTrx  `json:"dlg,omitempty"`
	CounterTrx `json:"ctr"`
	Compensate func(sg *Saga) error `json:"-"`
}

type SagaOrchestrator interface {
	IsActive() bool
	ExecuteSaga(ctx context.Context, saga *Saga) error
}
