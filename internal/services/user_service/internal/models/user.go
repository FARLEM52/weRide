package models

import (
	"github.com/google/uuid"
	"time"
)

type User struct {
	UserID    uuid.UUID
	Email     string
	PassHash  []byte
	FirsName  string
	LastName  string
	CreatedAt time.Time
}
