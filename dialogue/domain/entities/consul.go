package entities

type ConsulAgent interface {
	Register() error
	Unregister() error
}
