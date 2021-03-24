package entities

import (
	"context"
)

type Saga struct {
	Id          string               `json:"-"`
	MessagesIds []string             `json:"ids,omitempty"`
	Command     string               `json:"cmd,omitempty"`
	ChatId      string               `json:"cid,omitempty"`
	UserId      int                  `json:"uid,omitempty"`
	Num         uint                 `json:"num,omitempty"`
	Compensate  func(sg *Saga) error `json:"-"`
}

type SagaOrchestrator interface {
	IsActive() bool
	ExecuteSaga(ctx context.Context, saga *Saga) error
}
