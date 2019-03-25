package main

import (
	"github.com/google/uuid"
	"time"
)

type Session struct {
	ID        string `gorm:"type:varchar(36);primary_key"`
	UserID    uint
	CreatedAt time.Time `gorm:"index:created_at"`
	IP        string
	UserAgent string
	Hash      string
}

func (s *Session) BeforeCreate() (err error) {
	id, err := uuid.NewRandom()
	if err != nil {
		return err
	}
	s.ID = id.String()
	return
}
