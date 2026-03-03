package deltachat

import (
	"context"
	"io"
	"os"
	"os/exec"
	"sync"

	"github.com/creachadair/jrpc2"
	"github.com/creachadair/jrpc2/channel"
)

const deltachatRpcServerBin = "deltachat-rpc-server"

// IOTransport is a Delta Chat RPC transport using an external deltachat-rpc-server program.
type IOTransport struct {
	Stderr      io.Writer
	AccountsDir string
	Cmd         string
	cmd         *exec.Cmd
	stdin       io.WriteCloser
	client      *jrpc2.Client
	ctx         context.Context
	cancel      context.CancelFunc
	mu          sync.Mutex
}

// NewIOTransport creates a new IOTransport using the default deltachat-rpc-server binary.
func NewIOTransport() *IOTransport {
	return &IOTransport{Cmd: deltachatRpcServerBin, Stderr: os.Stderr}
}

// Open starts the deltachat-rpc-server process and connects to it.
func (trans *IOTransport) Open() error {
	trans.mu.Lock()
	defer trans.mu.Unlock()

	if trans.ctx != nil && trans.ctx.Err() == nil {
		return &TransportStartedErr{}
	}

	trans.ctx, trans.cancel = context.WithCancel(context.Background())
	trans.cmd = exec.CommandContext(trans.ctx, trans.Cmd)
	if trans.AccountsDir != "" {
		trans.cmd.Env = append(os.Environ(), "DC_ACCOUNTS_PATH="+trans.AccountsDir)
	}
	trans.cmd.Stderr = trans.Stderr
	var err error
	trans.stdin, err = trans.cmd.StdinPipe()
	if err != nil {
		trans.cancel()
		return err
	}
	stdout, err := trans.cmd.StdoutPipe()
	if err != nil {
		trans.cancel()
		return err
	}
	if err := trans.cmd.Start(); err != nil {
		trans.cancel()
		return err
	}

	trans.client = jrpc2.NewClient(channel.Line(stdout, trans.stdin), nil)
	return nil
}

// Close stops the deltachat-rpc-server process.
func (trans *IOTransport) Close() {
	trans.mu.Lock()
	defer trans.mu.Unlock()

	if trans.ctx == nil || trans.ctx.Err() != nil {
		return
	}

	_ = trans.stdin.Close()
	trans.cancel()
	trans.cmd.Wait() //nolint:errcheck
}

// Call requests the RPC server to call a function that does not have a return value.
func (trans *IOTransport) Call(ctx context.Context, method string, params ...any) error {
	_, err := trans.client.Call(ctx, method, params)
	return err
}

// CallResult requests the RPC server to call a function that does have a return value.
func (trans *IOTransport) CallResult(ctx context.Context, result any, method string, params ...any) error {
	return trans.client.CallResult(ctx, method, params, &result)
}

// TransportStartedErr is returned by IOTransport.Open() if the transport is already started.
type TransportStartedErr struct{}

func (e *TransportStartedErr) Error() string {
	return "transport is already started"
}
