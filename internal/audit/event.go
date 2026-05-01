package audit

import (
	"encoding/json"
	"time"
)

const (
	ActionShorten = "shorten"
	ActionFollow  = "follow"
)

type Event struct {
	TS          int64  `json:"ts"`
	Action      string `json:"action"`
	UserID      string `json:"user_id,omitempty"`
	OriginalURL string `json:"url"`
}

func NewEvent(action, userID, originalURL string) Event {
	return Event{
		TS:          time.Now().Unix(),
		Action:      action,
		UserID:      userID,
		OriginalURL: originalURL,
	}
}

func (e Event) MarshalJSONLine() ([]byte, error) {
	return json.Marshal(e)
}
