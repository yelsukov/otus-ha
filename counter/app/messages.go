package app

const (
	statusSuccess = 1
	statusAbort   = 2
	cmdIncr       = "incr"
	cmdDecr       = "decr"
)

type SagaOutboundMessage struct {
	SagaId string `json:"sid"`
	Action string `json:"act"`
	Status uint8  `json:"sts"`
}

type SagaInboundMessage struct {
	SagaId  string `json:"sid"`
	Command string `json:"cmd"`
	ChatId  string `json:"cid"`
	UserId  int    `json:"uid"`
	Num     uint   `json:"num"`
}
