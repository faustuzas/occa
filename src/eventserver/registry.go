package eventserver

import pkgid "github.com/faustuzas/occa/src/pkg/id"

type Registry interface {
	GetConnection(recipientID pkgid.ID) (Connection, error)
}
