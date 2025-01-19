package models

import "github.com/google/uuid"

type User struct {
	ID        uuid.UUID
	OAuth2ID  int
	Name      string
	AvatarUrl string
}
