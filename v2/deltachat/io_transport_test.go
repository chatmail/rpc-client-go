package deltachat

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestTransportStartedErr(t *testing.T) {
	t.Parallel()
	err := &TransportStartedErr{}
	require.NotEmpty(t, err.Error())
}

func TestIOTransport_OpenTwice(t *testing.T) {
	t.Parallel()
	acfactory.WithRpc(func(rpc *Rpc) {
		trans := rpc.Transport.(*IOTransport)
		err := trans.Open()
		require.NotNil(t, err)
		_, ok := err.(*TransportStartedErr)
		require.True(t, ok)
	})
}
