package storages

import (
	"github.com/yelsukov/otus-ha/news/domain/models"
)

type FollowerStorage interface {
	ReadOne(id int) (*models.Followers, error)
	ReadMany() ([]models.Followers, error)
	AddFollower(uid int, fid int) error
}
