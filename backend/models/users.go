package models

import (
	"github.com/yelsukov/otus-ha/backend/errors"
	"golang.org/x/crypto/bcrypt"
	"html"
	"strings"
	"time"
)

type User struct {
	Id           int64      `json:"id"`
	Username     string     `json:"username"`
	FirstName    NullString `json:"firstName"`
	LastName     NullString `json:"lastName"`
	Age          NullInt32  `json:"age"`
	Gender       string     `json:"gender"`
	City         NullString `json:"city"`
	Interests    NullString `json:"interests"`
	PasswordHash []byte     `json:"-"`
	IsFriend     bool       `json:"isFriend,omitempty"`
	CreatedAt    time.Time  `json:"createdAt"`
}

type UserList struct {
	Object string `json:"object"`
}

func (u *User) Validate() error {
	if u.Username != "" {
		ul := len(u.Username)
		if ul < 5 {
			return errors.New("4001", "`Username` can't be less than 5 chars")
		}
		if ul > 30 {
			return errors.New("4001", "`Username` can't be greater then 30 chars")
		}
	}

	if u.FirstName.Valid && len(u.FirstName.String) > 20 {
		return errors.New("4003", "`First Name` can't contain more than 20 chars")
	}
	if u.LastName.Valid && len(u.LastName.String) > 30 {
		return errors.New("4004", "`Last Name` can't contain more than 30 chars")
	}
	if u.Age.Valid && (u.Age.Int32 < 18 || u.Age.Int32 > 120) {
		return errors.New("4005", "`Age` should be between 18 and 120")
	}
	if u.Gender != "f" && u.Gender != "m" {
		return errors.New("4006", "Invalid format of `Gender`")
	}
	if u.City.Valid && len(u.City.String) > 80 {
		return errors.New("4007", "`City` can't be greater than 80 chars")
	}
	if u.Interests.Valid && len(u.Interests.String) > 500 {
		return errors.New("4008", "`Interests` can't be greater than 500 chars")
	}

	return nil
}

func (u *User) Sanitize() {
	if u.Username != "" {
		u.Username = html.EscapeString(strings.TrimSpace(u.Username))
	}
	if u.FirstName.Valid && u.FirstName.String != "" {
		u.FirstName.String = html.EscapeString(strings.TrimSpace(u.FirstName.String))
	}
	if u.LastName.Valid && u.LastName.String != "" {
		u.LastName.String = html.EscapeString(strings.TrimSpace(u.LastName.String))
	}
	if u.City.Valid {
		u.City.String = html.EscapeString(strings.TrimSpace(u.City.String))
	}
	if u.Interests.Valid {
		u.Interests.String = html.EscapeString(strings.TrimSpace(u.Interests.String))
	}
}

func (u *User) SetPassword(password string) error {
	var err error
	u.PasswordHash, err = bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return err
}
