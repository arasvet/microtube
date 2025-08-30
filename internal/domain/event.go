package domain

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

type EventType string

const (
	EventViewStart    EventType = "view_start"
	EventViewComplete EventType = "view_complete"
	EventLike         EventType = "like"
	EventSearchQuery  EventType = "search_query"
	EventClickResult  EventType = "click_result"
)

var ErrInvalidEvent = errors.New("invalid event")

type Event struct {
	EventID   uuid.UUID
	TS        time.Time
	Type      EventType
	SessionID string
	UserID    uuid.UUID
	VideoID   uuid.UUID
	Query     string
	DwellMs   int
}

func (e *Event) Validate() error {
	if e.EventID == uuid.Nil || e.SessionID == "" || e.Type == "" || e.TS.IsZero() {
		return ErrInvalidEvent
	}
	switch e.Type {
	case EventViewStart, EventViewComplete, EventLike:
		if e.VideoID.String() == "" {
			return errors.New("video_id required for view/like")
		}
	case EventSearchQuery:
		if e.Query == "" {
			return errors.New("query required for search_query")
		}
	case EventClickResult:
		if e.VideoID.String() == "" || e.Query == "" {
			return errors.New("video_id and query required for click_result")
		}
	default:
		return ErrInvalidEvent
	}
	return nil
}
