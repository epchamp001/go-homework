package models

import (
	"github.com/google/uuid"
	"time"
)

type OutboxRecord struct {
	ID            uuid.UUID
	Payload       []byte
	Attempts      int
	LastAttemptAt *time.Time
	CreatedAt     time.Time
}
