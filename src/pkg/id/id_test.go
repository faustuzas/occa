package id

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestIDJsonRoundTrip(t *testing.T) {
	id := FromString("e8eeb523-2cc9-4f24-b878-0ad97aef5c90")

	var payload = struct {
		ID ID `json:"id"`
	}{ID: id}

	marshalledBytes, err := json.Marshal(payload)
	require.NoError(t, err)
	require.Equal(t, `{"id":"e8eeb523-2cc9-4f24-b878-0ad97aef5c90"}`, string(marshalledBytes))

	payload.ID = ID{}
	require.NoError(t, json.Unmarshal(marshalledBytes, &payload))
	require.Equal(t, id, payload.ID)
}
