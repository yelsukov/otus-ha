package requests

import "github.com/yelsukov/otus-ha/backend/models"

type SignRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type UpdateProfileRequest struct {
	models.User
}
