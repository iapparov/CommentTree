package app

import (
	"errors"
	"github.com/google/uuid"
	wbzlog "github.com/wb-go/wbf/zlog"
	"time"
)

// TODO: add dto, entity separation
type Comment struct {
	ID        uuid.UUID  `json:"id"`
	Text      string     `json:"text"`
	CreatedAt time.Time  `json:"created_at"`
	ParentID  *uuid.UUID `json:"parent_id"`
}

func NewComment(parentid string, text string) (*Comment, error) {
	var c Comment
	if parentid != "" {
		parentuuid, err := uuid.Parse(parentid)
		if err != nil {
			wbzlog.Logger.Error().Err(err).Msg("bad parent id")
			return nil, err
		}
		c.ParentID = &parentuuid
	} else {
		c.ParentID = nil
	}
	if text == "" {
		err := errors.New("text is empty")
		wbzlog.Logger.Error().Err(err)
		return nil, err
	}
	c.Text = text
	c.ID = uuid.New()
	c.CreatedAt = time.Now()
	return &c, nil
}
