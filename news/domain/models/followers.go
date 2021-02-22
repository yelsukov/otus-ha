package models

type Followers struct {
	Id   int   `bson:"_id"`
	List []int `bson:"lst"`
}
