package storages

import (
	"database/sql"
	"github.com/yelsukov/otus-ha/backend/errors"
)

type FriendsStorage struct {
	db *sql.DB
}

func NewFriendsStorage(db *sql.DB) *FriendsStorage {
	return &FriendsStorage{db}
}

func (m *FriendsStorage) IsFriend(userId int64, friendId int64) (exists bool, err error) {
	err = m.db.QueryRow("SELECT EXISTS(SELECT 1 FROM `friends` WHERE `user_id` = ? AND `friend_id` = ?)", userId, friendId).Scan(&exists)
	return
}

func (m *FriendsStorage) Add(userId int64, friendId int64) error {
	if userId == friendId {
		return errors.New("4008", "you are your own best friend, but not on today")
	}

	exists, err := m.IsFriend(userId, friendId)
	if err != nil {
		return err
	}
	if exists {
		// Already have an old friendship
		return nil
	}

	stmt, err := m.db.Prepare("INSERT INTO `friends`(`user_id`, `friend_id`) VALUES(?,?),(?,?)")
	defer closeStmt(stmt)
	if err != nil {
		return err
	}
	_, err = stmt.Exec(userId, friendId, friendId, userId)

	return err
}

func (m *FriendsStorage) Delete(userId int64, friendId int64) error {
	stmt, err := m.db.Prepare("DELETE FROM `friends` WHERE `user_id` IN (?,?) AND `friend_id` IN(?,?)")
	defer closeStmt(stmt)
	if err != nil {
		return err
	}

	_, err = stmt.Exec(userId, friendId, userId, friendId)

	return err
}
