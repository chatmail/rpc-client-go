// RPC API definitions
package main

import (
	"github.com/chatmail/rpc-client-go/deltachat/xdcrpc"
)

// You must put your available RPC API in a custom type
type API struct{}

// Function without arguments or return value
func (api *API) Noop() {
	// do nothing
}

// Function with return value but no *xdcrpc.Error
func (api *API) Echo(text string) string {
	return text
}

// Function that might return an xdcrpc.Error.
// Functions must return `*xdcrpc.Error` instead of `error`
func (api *API) Divide(a int, b int) (int, *xdcrpc.Error) {
	if b == 0 {
		return 0, &xdcrpc.Error{Code: 1, Message: "Division by zero"}
	}
	return a / b, nil
}
