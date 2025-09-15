module rpcbot

go 1.25

require github.com/chatmail/rpc-client-go v1.134.0

require (
	github.com/creachadair/jrpc2 v1.1.2 // indirect
	github.com/creachadair/mds v0.8.2 // indirect
	golang.org/x/sync v0.6.0 // indirect
)

// this is needed only for tests, don't add it in your project's go.mod
replace github.com/chatmail/rpc-client-go => ../../
