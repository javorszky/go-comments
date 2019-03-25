package main

import (
	"time"

	"github.com/google/uuid"
)

// Session struct holds the model for the user sessions. Users have many sessions. Sessions belong to one user.
type Session struct {
	ID        string `gorm:"type:varchar(36);primary_key"`
	UserID    uint
	CreatedAt time.Time `gorm:"index:created_at"`
	IP        string
	UserAgent string
	Hash      string
}

// BeforeCreate is a hook function gorm uses. We create a uuidv4 as an ID for the model.
func (s *Session) BeforeCreate() (err error) {
	id, err := uuid.NewRandom()
	if err != nil {
		return err
	}
	s.ID = id.String()
	return
}
