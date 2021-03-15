package jwt

import (
	"errors"
	"github.com/dgrijalva/jwt-go"
	"time"
)

type Tokenizer interface {
	Tokenize(userId int64, username string) (*Token, error)
	ExtractUserId(token string) (int64, error)
}

type JWT struct {
	Secret []byte
	Ttl    time.Duration
}

type Token struct {
	Object    string `json:"object"`
	Token     string `json:"token"`
	UserId    int64  `json:"userId"`
	Username  string `json:"username"`
	ExpiresAt int64  `json:"expiresAt"`
}

type claims struct {
	UserId int64 `json:"user_id"`
	jwt.StandardClaims
}

func (j *JWT) Tokenize(userId int64, username string) (*Token, error) {
	// Declare the expiration time of the token
	issuedAt := time.Now()
	expirationTime := issuedAt.Add(j.Ttl).Unix()

	// Declare the token with the algorithm used for signing, and the claims
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, &claims{
		UserId: userId,
		StandardClaims: jwt.StandardClaims{
			Issuer:    "ha-backend",
			IssuedAt:  issuedAt.Unix(),
			ExpiresAt: expirationTime,
		},
	})
	// Create the JWT string
	strJwt, err := token.SignedString(j.Secret)
	if err != nil {
		return nil, err
	}

	return &Token{
		"token",
		strJwt,
		userId,
		username,
		expirationTime,
	}, nil
}

func (j *JWT) ExtractUserId(tokenStr string) (int64, error) {
	// Initialize a new instance of `claims`
	clm := &claims{}

	token, err := jwt.ParseWithClaims(tokenStr, clm, func(token *jwt.Token) (interface{}, error) {
		return j.Secret, nil
	})
	if err != nil {
		return 0, errors.New("invalid authorization token")
	}
	if !token.Valid {
		return 0, errors.New("authorization token is expired")
	}

	if clm.UserId == 0 {
		return 0, errors.New("invalid authorization token")
	}

	return clm.UserId, nil
}
