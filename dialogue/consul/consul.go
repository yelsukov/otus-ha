package consul

import (
	"github.com/hashicorp/consul/api"
	"github.com/yelsukov/otus-ha/dialogue/config"
	"strconv"
)

type Agent struct {
	agent *api.Agent
	cfg   *config.Config
}

func NewAgent(cfg *config.Config) (*Agent, error) {
	consulCfg := api.DefaultConfig()
	consulCfg.Address = cfg.ConsulDsn
	client, err := api.NewClient(consulCfg)
	if err != nil {
		return nil, err
	}

	return &Agent{client.Agent(), cfg}, nil
}

func (a *Agent) Register() error {
	p, err := strconv.Atoi(a.cfg.ServicePort)
	if err != nil {
		return err
	}

	err = a.agent.ServiceRegister(&api.AgentServiceRegistration{
		ID:      a.cfg.ServiceId,
		Name:    a.cfg.ServiceName,
		Address: a.cfg.ServiceHost,
		Port:    p,
		Check: &api.AgentServiceCheck{
			Interval: "5s",
			Timeout:  "2s",
			HTTP:     "http://" + a.cfg.ServiceHost + ":" + a.cfg.ServicePort + "/ping",
		},
	})

	return err
}

func (a *Agent) Unregister() error {
	return a.agent.ServiceDeregister(a.cfg.ServiceId)
}
