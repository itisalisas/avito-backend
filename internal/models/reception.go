package models

import (
	"github.com/google/uuid"
	"time"
)

type Status string

const (
	StatusInProgress Status = "in_progress"
	StatusCLose      Status = "close"
)

type Reception struct {
	ID       uuid.UUID `json:"id"`
	DateTime time.Time `json:"dateTime"`
	PvzID    uuid.UUID `json:"pvzId"`
	Status   Status    `json:"status"`
}
