package eventserver

import (
	"context"
	"time"

	pkgid "github.com/faustuzas/occa/src/pkg/id"
)

type Event struct {
	SenderID         pkgid.ID  `json:"senderID"`
	RecipientID      pkgid.ID  `json:"recipientID"`
	Content          string    `json:"content"`
	SentFromClientAt time.Time `json:"sentFromClientAt"`
}

type Connection interface {
	SendEvent(ctx context.Context, msg Event) error
}
