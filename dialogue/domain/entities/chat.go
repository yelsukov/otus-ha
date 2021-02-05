package entities

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
	"sort"
	"time"
)

type Chat struct {
	Id        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Users     []int              `bson:"users" json:"users"`
	CreatedAt int64              `bson:"ts" json:"created_at"`
}

func (c *Chat) Validate() error {
	unique(&c.Users)
	if len(c.Users) < 2 {
		return NewError("4004", "it should at least 2 uniq users")
	}
	return nil
}

func (c *Chat) BeforeSave() {
	c.CreatedAt = time.Now().Unix()
}

// Unique removes duplicate elements from data. It assumes sort.IsSorted(data).
func unique(data *[]int) {
	sort.Ints(*data)
	n := len(*data)
	k := 0
	if n != 0 {
		for i := 1; i < n; i++ {
			if (*data)[k] < (*data)[i] {
				k++
				(*data)[k], (*data)[i] = (*data)[i], (*data)[k]
			}
		}
	}
	*data = (*data)[:k+1]
}
