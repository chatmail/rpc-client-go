package deltachat

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPair_UnmarshalJSON(t *testing.T) {
	t.Parallel()
	var p Pair[string, int]
	require.Nil(t, json.Unmarshal([]byte(`["hello", 42]`), &p))
	require.Equal(t, "hello", p.First)
	require.Equal(t, 42, p.Second)

	require.NotNil(t, json.Unmarshal([]byte(`"notarray"`), &p))
	require.NotNil(t, json.Unmarshal([]byte(`[1, 2]`), &p))
	require.NotNil(t, json.Unmarshal([]byte(`["hello", "notint"]`), &p))
}

func TestUnmarshalAccount(t *testing.T) {
	t.Parallel()

	var acc Account
	require.Nil(t, unmarshalAccount(json.RawMessage(`{"kind":"Configured","id":1,"color":"#fff"}`), &acc))
	require.Equal(t, "Configured", acc.GetKind())

	require.Nil(t, unmarshalAccount(json.RawMessage(`{"kind":"Unconfigured","id":2}`), &acc))
	require.Equal(t, "Unconfigured", acc.GetKind())

	require.NotNil(t, unmarshalAccount(json.RawMessage(`{"kind":"Unknown"}`), &acc))
	require.NotNil(t, unmarshalAccount(json.RawMessage(`notjson`), &acc))
}
