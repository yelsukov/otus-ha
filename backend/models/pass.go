package models

import (
	"github.com/yelsukov/otus-ha/backend/errors"
	"strconv"
	"unicode/utf8"
)

type PassChange struct {
	Password        string
	PasswordConfirm string
}

const (
	minPasswordLength = 10
	maxPasswordLength = 64
)

func (p *PassChange) Validate() error {
	if p.Password == "" {
		return errors.New("4002", "`Password` is Required")
	}
	if p.Password != p.PasswordConfirm {
		return errors.New("4002", "`Password` differs from Repeated")
	}
	passwordLength := utf8.RuneCountInString(p.Password)
	if passwordLength < minPasswordLength || passwordLength > maxPasswordLength {
		return errors.New("4002", "`Password` should be at least "+strconv.Itoa(minPasswordLength)+" chars, and not longer than "+strconv.Itoa(maxPasswordLength)+" symbols")
	}
	return nil
}
