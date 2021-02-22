package entities

type CacheHeater interface {
	HeatFollowers()
	HeatEvents(uid int)
}
