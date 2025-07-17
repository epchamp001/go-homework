package models

import (
	"time"

	"github.com/google/uuid"
)

type OutboxRecord struct {
	ID            uuid.UUID
	Payload       []byte
	Attempts      int
	LastAttemptAt *time.Time
	CreatedAt     time.Time
}
