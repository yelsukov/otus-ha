package queues

const (
	StatusSuccess = 1
	StatusAbort   = 2
)

type SagaInboundMessage struct {
	SagaId string `json:"sid"`
	Action string `json:"act"`
	Status uint8  `json:"sts"`
}

type SagaCounterMessage struct {
	SagaId  string `json:"sid"`
	Command string `json:"cmd"`
	ChatId  string `json:"cid"`
	UserId  int    `json:"uid"`
	Num     uint   `json:"num"`
}
