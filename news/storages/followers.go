package storages

import (
	"context"
	log "github.com/sirupsen/logrus"
	"github.com/yelsukov/otus-ha/news/domain/entities"
	"github.com/yelsukov/otus-ha/news/domain/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type FollowerStorage struct {
	ctx   context.Context
	col   *mongo.Collection
	cache entities.Cache
}

func NewFollowerStorage(ctx context.Context, db *mongo.Database, cache entities.Cache) *FollowerStorage {
	return &FollowerStorage{ctx, db.Collection("followers"), cache}
}

func (f *FollowerStorage) ReadOne(uid int) (*models.Followers, error) {
	list, err := f.cache.ReadFollowers(uid)
	if err != nil {
		log.WithError(err).Error("fail on reading followers from cache")
	}
	// found in cache
	if len(list) != 0 {
		log.Debugf("got %d followers from Cache for user #%d", len(list), uid)
		return &models.Followers{Id: uid, List: list}, nil
	}

	// not found in cache, trying to get from db
	var followers models.Followers
	err = f.col.FindOne(f.ctx, bson.M{"_id": uid}).Decode(&followers)
	if err != nil && err != mongo.ErrNoDocuments {
		return nil, err
	}
	log.Debugf("got %d followers from DB for user #%d", len(followers.List), uid)
	// write to cache
	if len(followers.List) != 0 {
		if err := f.cache.AddFollowers(uid, followers.List...); err != nil {
			log.WithError(err).Error("fail on caching followers list")
		}
	}

	return &followers, nil
}

func (f *FollowerStorage) ReadMany() ([]models.Followers, error) {
	var followers []models.Followers

	cursor, err := f.col.Find(f.ctx, bson.D{})
	if err != nil {
		return nil, err
	}
	if err = cursor.All(f.ctx, &followers); err != nil {
		return nil, err
	}

	return followers, nil
}

func (f *FollowerStorage) AddFollower(uid int, fid int) error {
	_, err := f.col.UpdateOne(
		f.ctx,
		bson.M{"_id": uid},
		bson.D{
			{"$push", bson.D{{"list", fid}}},
		},
		options.Update().SetUpsert(true),
	)
	return err
}
