package serviceboot

import (
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/faustuzas/occa/src/eventserver"
	pkgnet "github.com/faustuzas/occa/src/pkg/net"
	pkgtest "github.com/faustuzas/occa/src/pkg/test"
)

func DefaultEventServerParams(t *testing.T) eventserver.Params {
	httpListenAddr := pkgnet.ListenAddrFromAddress("0.0.0.0:0")
	httpListener, err := httpListenAddr.Listener()
	require.NoError(t, err)

	t.Cleanup(func() {
		_ = httpListener.Close()
	})

	grpcListenAddr := pkgnet.ListenAddrFromAddress("0.0.0.0:0")
	grpcListener, err := grpcListenAddr.Listener()
	require.NoError(t, err)

	t.Cleanup(func() {
		_ = grpcListener.Close()
	})

	closeCh := make(chan struct{})
	t.Cleanup(func() {
		close(closeCh)
	})

	return eventserver.Params{
		Configuration: eventserver.Configuration{
			HTTPListenAddress: httpListenAddr,
			GRPCListenAddress: grpcListenAddr,
		},
		Logger:  pkgtest.Logger.With(zap.String("component", "chat-server")),
		CloseCh: closeCh,
	}
}
