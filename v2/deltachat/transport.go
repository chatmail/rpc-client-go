package deltachat

import (
	"context"
)

// RpcTransport is the Delta Chat RPC client's transport interface.
type RpcTransport interface {
	// Call requests the RPC server to call a function that does not have a return value.
	Call(ctx context.Context, method string, params ...any) error
	// CallResult requests the RPC server to call a function that does have a return value.
	CallResult(ctx context.Context, result any, method string, params ...any) error
}
