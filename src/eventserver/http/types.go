package http

import pkgid "github.com/faustuzas/occa/src/pkg/id"

type SendMessageRequest struct {
	RecipientID pkgid.ID `json:"recipientID"`
	Content     string   `json:"content"`
}
