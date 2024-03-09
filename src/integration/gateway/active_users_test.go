package gateway

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/faustuzas/occa/src/gateway"
)

func TestGateway_ActiveUsers(t *testing.T) {
	params := DefaultParams(t)
	go func() {
		require.NoError(t, gateway.Start(params))
	}()

	user1 := NewTester(t, context.Background(), params.HTTPListenAddress.String())
	user1.Register("user_1", "password")
	user1.Login("user_1", "password")

	activeUsers := user1.ActiveUsers().ActiveUsers
	require.Empty(t, activeUsers)

	user1.Heartbeat()

	activeUsers = user1.ActiveUsers().ActiveUsers
	require.Len(t, activeUsers, 1)
	require.Equal(t, "user_1", activeUsers[0].Username)
	require.NotEmpty(t, activeUsers[0].ID)
	require.NotEmpty(t, activeUsers[0].LastSeen)
}
