package transport

import (
	"context"
	"io"
	"os"
	"os/exec"
	"sync"

	"github.com/creachadair/jrpc2"
	"github.com/creachadair/jrpc2/channel"
)

// Delta Chat RPC transport using external deltachat-rpc-server program
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

func NewIOTransport() *IOTransport {
	return &IOTransport{Cmd: deltachatRpcServerBin, Stderr: os.Stderr}
}

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
	trans.stdin, _ = trans.cmd.StdinPipe()
	stdout, _ := trans.cmd.StdoutPipe()
	if err := trans.cmd.Start(); err != nil {
		trans.cancel()
		return err
	}

	trans.client = jrpc2.NewClient(channel.Line(stdout, trans.stdin), nil)
	return nil
}

func (trans *IOTransport) Close() {
	trans.mu.Lock()
	defer trans.mu.Unlock()

	if trans.ctx == nil || trans.ctx.Err() != nil {
		return
	}

	_ = trans.stdin.Close()
	trans.cancel()
	trans.cmd.Process.Wait() //nolint:errcheck
}

func (trans *IOTransport) Call(ctx context.Context, method string, params ...any) error {
	_, err := trans.client.Call(ctx, method, params)
	return err
}

func (trans *IOTransport) CallResult(ctx context.Context, result any, method string, params ...any) error {
	return trans.client.CallResult(ctx, method, params, &result)
}

// TransportStartedErr is returned by IOTransport.Open() if the Transport is already started
type TransportStartedErr struct{}

func (trans *TransportStartedErr) Error() string {
	return "transport is already started"
}
