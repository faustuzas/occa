package rteventspb

import pkgid "github.com/faustuzas/occa/src/pkg/id"

func NewDirectMessageEvent(senderID pkgid.ID, message string) *Event {
	return &Event{
		Payload: &Event_DirectMessage{
			DirectMessage: &DirectMessage{
				SenderId: senderID.String(),
				Message:  message,
			},
		},
	}
}
