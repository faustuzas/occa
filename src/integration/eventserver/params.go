package eventserver

import (
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/faustuzas/occa/src/eventserver"
	"github.com/faustuzas/occa/src/integration/containers"
	pkgauth "github.com/faustuzas/occa/src/pkg/auth"
	pkgetcd "github.com/faustuzas/occa/src/pkg/etcd"
	pkgmemstore "github.com/faustuzas/occa/src/pkg/memstore"
	pkgnet "github.com/faustuzas/occa/src/pkg/net"
	pkgtest "github.com/faustuzas/occa/src/pkg/test"
)

func DefaultParams(t *testing.T) eventserver.Params {
	var (
		etcd  = containers.WithEtcd(t)
		redis = containers.WithRedis(t)
	)

	httpListenAddr := pkgnet.ListenAddrFromAddress("0.0.0.0:0")
	_, err := httpListenAddr.Listener()
	require.NoError(t, err)

	grpcListenAddr := pkgnet.ListenAddrFromAddress("0.0.0.0:0")
	_, err = grpcListenAddr.Listener()
	require.NoError(t, err)

	closeCh := make(chan struct{})
	t.Cleanup(func() {
		close(closeCh)
	})

	pubKey, _, err := pkgtest.GetRSAPairPaths()
	require.NoError(t, err)

	return eventserver.Params{
		Configuration: eventserver.Configuration{
			ServerID:          "test_server_1",
			HTTPListenAddress: httpListenAddr,
			GRPCListenAddress: grpcListenAddr,
			Etcd: pkgetcd.Configuration{
				Username:  etcd.Username,
				Password:  etcd.Password,
				Endpoints: []string{fmt.Sprintf("http://localhost:%d", etcd.Port)},
			},
			MemStore: pkgmemstore.Configuration{
				User:     redis.Username,
				Password: redis.Password,
				Address:  fmt.Sprintf("localhost:%d", redis.Port),
				Prefix:   uuid.New().String(),
			},
			Auth: pkgauth.ValidatorConfiguration{
				Type: pkgauth.ValidatorConfigurationJWTRSA,
				JWTValidator: pkgauth.JWTValidatorConfiguration{
					PublicKeyPath: pubKey,
				},
			},
		},
		Logger:  pkgtest.Instrumentation.Logger.With(zap.String("component", "chat-server"), zap.String("test", t.Name())),
		CloseCh: closeCh,
	}
}
