package deltachat

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestRpc_ServerSchema(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	schemaJSON, err := exec.CommandContext(ctx, deltachatRpcServerBin, "--openrpc").Output()
	require.NoError(t, err)
	var schema struct {
		Methods []struct {
			Name string `json:"name"`
		} `json:"methods"`
	}
	require.NoError(t, json.Unmarshal(schemaJSON, &schema))
	require.NotEmpty(t, schema.Methods)

	serverMethods := make([]string, 0, len(schema.Methods))
	for _, method := range schema.Methods {
		serverMethods = append(serverMethods, strings.ReplaceAll(method.Name, "_", ""))
	}
	rpcType := reflect.TypeOf(&Rpc{})
	clientMethods := make([]string, 0, rpcType.NumMethod())
	for index := 0; index < rpcType.NumMethod(); index++ {
		clientMethods = append(clientMethods, strings.ToLower(rpcType.Method(index).Name))
	}
	require.ElementsMatch(t, serverMethods, clientMethods)
}

func TestRpc_Server262Requests(t *testing.T) {
	t.Parallel()
	qr := "dcaccount:example.invalid"
	tests := []struct {
		name     string
		method   string
		params   []any
		response string
		want     any
		call     func(*Rpc) (any, error)
	}{
		{"initialize transports", "init_transports", []any{uint32(7), &qr}, "", nil,
			func(rpc *Rpc) (any, error) { return nil, rpc.InitTransports(7, &qr) }},
		{"initialize default transports", "init_transports", []any{uint32(7), (*string)(nil)}, "", nil,
			func(rpc *Rpc) (any, error) { return nil, rpc.InitTransports(7, nil) }},
		{"pin message", "set_pinned_message_state", []any{uint32(7), uint32(42), true}, "", nil,
			func(rpc *Rpc) (any, error) { return nil, rpc.SetPinnedMessageState(7, 42, true) }},
		{"unpin message", "set_pinned_message_state", []any{uint32(7), uint32(42), false}, "", nil,
			func(rpc *Rpc) (any, error) { return nil, rpc.SetPinnedMessageState(7, 42, false) }},
		{"pinned messages", "get_pinned_messages", []any{uint32(7), uint32(9)}, "[42,43]", []uint32{42, 43},
			func(rpc *Rpc) (any, error) { return rpc.GetPinnedMessages(7, 9) }},
		{"wait for work", "wait_for_all_work_done", []any{uint32(7)}, "", nil,
			func(rpc *Rpc) (any, error) { return nil, rpc.WaitForAllWorkDone(7) }},
		{"app version", "get_app_version", []any{"client", "source"},
			`{"downloadUrl":"https://example.invalid/app","versionInteger":62,"versionString":"2.62.0"}`,
			&AppSource{DownloadUrl: "https://example.invalid/app", VersionInteger: 62, VersionString: "2.62.0"},
			func(rpc *Rpc) (any, error) { return rpc.GetAppVersion("client", "source") }},
		{"unknown app version", "get_app_version", []any{"client", "source"}, "null", (*AppSource)(nil),
			func(rpc *Rpc) (any, error) { return rpc.GetAppVersion("client", "source") }},
		{"sending finished", "is_sending_finished", nil, "true", true,
			func(rpc *Rpc) (any, error) { return rpc.IsSendingFinished() }},
		{"sending pending", "is_sending_finished", nil, "false", false,
			func(rpc *Rpc) (any, error) { return rpc.IsSendingFinished() }},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			transport := &server262Transport{t: t, ctx: ctx, method: test.method, params: test.params, response: test.response}
			rpc := &Rpc{Context: ctx, Transport: transport}
			got, err := test.call(rpc)
			require.NoError(t, err)
			require.Equal(t, test.want, got)

			transport.err = errors.New("server request failed")
			_, err = test.call(rpc)
			require.ErrorIs(t, err, transport.err)
		})
	}
}

type server262Transport struct {
	t        *testing.T
	ctx      context.Context
	method   string
	params   []any
	response string
	err      error
}

func (transport *server262Transport) Call(ctx context.Context, method string, params ...any) error {
	transport.t.Helper()
	require.Equal(transport.t, transport.ctx, ctx)
	require.Equal(transport.t, transport.method, method)
	require.Equal(transport.t, transport.params, params)
	return transport.err
}

func (transport *server262Transport) CallResult(ctx context.Context, destination any, method string, params ...any) error {
	transport.t.Helper()
	if err := transport.Call(ctx, method, params...); err != nil {
		return err
	}
	return json.Unmarshal([]byte(transport.response), destination)
}

func TestServer262Freshness(t *testing.T) {
	t.Parallel()
	for _, freshness := range []ContactFreshness{ContactFreshnessNormal, ContactFreshnessRecentlySeen, ContactFreshnessOld} {
		t.Run(string(freshness), func(t *testing.T) {
			payload := []byte(`{"freshness":"` + string(freshness) + `"}`)
			var contact Contact
			require.NoError(t, json.Unmarshal(payload, &contact))
			require.Equal(t, freshness, contact.Freshness)
			var chat FullChat
			require.NoError(t, json.Unmarshal(payload, &chat))
			require.Equal(t, freshness, chat.Freshness)
			var item ChatListItemFetchResult
			require.NoError(t, unmarshalChatListItemFetchResult(json.RawMessage(`{"kind":"ChatListItem","freshness":"`+string(freshness)+`"}`), &item))
			require.Equal(t, freshness, item.(*ChatListItemFetchResultChatListItem).Freshness)
		})
	}
}

func TestServer262PinnedMessageDecoding(t *testing.T) {
	t.Parallel()
	payload := []byte(`{"kind":"message","id":42,"isPinned":true,"systemMessageType":"MessagePinned","sender":{"freshness":"RecentlySeen"}}`)
	var message Message
	require.NoError(t, json.Unmarshal(payload, &message))
	require.True(t, message.IsPinned)
	require.Equal(t, SystemMessageTypeMessagePinned, message.SystemMessageType)
	require.Equal(t, ContactFreshnessRecentlySeen, message.Sender.Freshness)

	var loaded MessageLoadResult
	require.NoError(t, unmarshalMessageLoadResult(payload, &loaded))
	pinned := loaded.(*MessageLoadResultMessage)
	require.True(t, pinned.IsPinned)
	require.Equal(t, ContactFreshnessRecentlySeen, pinned.Sender.Freshness)
	encoded, err := json.Marshal(pinned)
	require.NoError(t, err)
	require.Contains(t, string(encoded), `"isPinned":true`)

	require.NoError(t, json.Unmarshal([]byte(`{"isPinned":false,"systemMessageType":"MessageUnpinned"}`), &message))
	require.False(t, message.IsPinned)
	require.Equal(t, SystemMessageTypeMessageUnpinned, message.SystemMessageType)
}

func TestServer262LocalAccount(t *testing.T) {
	t.Parallel()
	transport := NewIOTransport()
	transport.AccountsDir = t.TempDir()
	require.NoError(t, transport.Open())
	t.Cleanup(transport.Close)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	rpc := &Rpc{Context: ctx, Transport: transport}
	accountID, err := rpc.AddAccount()
	require.NoError(t, err)
	contactID, err := rpc.CreateContact(accountID, "local@example.invalid", nil)
	require.NoError(t, err)
	contact, err := rpc.GetContact(accountID, contactID)
	require.NoError(t, err)
	require.Equal(t, ContactFreshnessNormal, contact.Freshness)

	// Invalid account IDs exercise these methods without creating remote accounts or starting mail I/O.
	require.ErrorContains(t, rpc.InitTransports(0, nil), "account")
	require.ErrorContains(t, rpc.WaitForAllWorkDone(0), "account")

	chatID, err := rpc.CreateChatByContactId(accountID, contactID)
	require.NoError(t, err)
	pins, err := rpc.GetPinnedMessages(accountID, chatID)
	require.NoError(t, err)
	require.Empty(t, pins)
	require.Error(t, rpc.SetPinnedMessageState(accountID, 0, true))
	chat, err := rpc.GetFullChatById(accountID, chatID)
	require.NoError(t, err)
	require.Equal(t, ContactFreshnessNormal, chat.Freshness)
}

func TestRunTestsServerVersion(t *testing.T) {
	t.Parallel()
	root, err := filepath.Abs("../..")
	require.NoError(t, err)
	for _, test := range []struct {
		name       string
		server     string
		wantExit   int
		wantOutput string
	}{
		{"matching stderr", "echo 2.62.0 >&2", 17, "version accepted"},
		{"matching stdout", "echo 2.62.0", 17, "version accepted"},
		{"older", "echo 2.56.0", 1, "Expected deltachat-rpc-server 2.62.0, found 2.56.0"},
		{"newer", "echo 2.63.0", 1, "Expected deltachat-rpc-server 2.62.0, found 2.63.0"},
		{"broken", "echo 2.62.0 >&2; exit 23", 23, "Checking code with gofmt"},
	} {
		t.Run(test.name, func(t *testing.T) {
			bin := t.TempDir()
			for name, body := range map[string]string{
				"gofmt":                "exit 0",
				"deltachat-rpc-server": test.server,
				"golangci-lint":        "echo 'version accepted'; exit 17",
			} {
				require.NoError(t, os.WriteFile(filepath.Join(bin, name), []byte("#!/bin/sh\n"+body+"\n"), 0o700))
			}
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()
			command := exec.CommandContext(ctx, "bash", "scripts/run_tests.sh")
			command.Dir = root
			command.Env = append(os.Environ(), "PATH="+bin)
			output, err := command.CombinedOutput()
			var exitError *exec.ExitError
			require.ErrorAs(t, err, &exitError)
			require.Equal(t, test.wantExit, exitError.ExitCode(), string(output))
			require.Contains(t, string(output), test.wantOutput)
			if test.wantExit != 17 {
				require.NotContains(t, string(output), "version accepted")
			}
		})
	}
}

func TestRpc_InitTransports(t *testing.T) {
	t.Parallel()
	acfactory.WithUnconfiguredAccount(func(rpc *Rpc, accountID uint32) {
		ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
		defer cancel()
		client := &Rpc{Context: ctx, Transport: rpc.Transport}

		configured, err := client.IsConfigured(accountID)
		require.NoError(t, err)
		require.False(t, configured)
		transports, err := client.ListTransports(accountID)
		require.NoError(t, err)
		require.Zero(t, len(transports))

		// An explicit setup QR confines account creation to the test relay.
		require.NoError(t, client.InitTransports(accountID, &acfactory.ConfigQr))
		configured, err = client.IsConfigured(accountID)
		require.NoError(t, err)
		require.True(t, configured)
		transports, err = client.ListTransports(accountID)
		require.NoError(t, err)
		require.Equal(t, 1, len(transports))
		originalAddress := transports[0].Addr
		require.NotEmpty(t, originalAddress)

		// On a configured profile, nil must be a no-op, not automatic relay selection.
		require.NoError(t, client.InitTransports(accountID, nil))
		transports, err = client.ListTransports(accountID)
		require.NoError(t, err)
		require.Equal(t, 1, len(transports))
		require.True(t, originalAddress == transports[0].Addr, "initial transport was replaced")
		configured, err = client.IsConfigured(accountID)
		require.NoError(t, err)
		require.True(t, configured)
	})
}

func TestRpc_WaitForAllWorkDone(t *testing.T) {
	t.Parallel()
	acfactory.WithUnconfiguredAccount(func(rpc *Rpc, accountID uint32) {
		ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
		defer cancel()
		client := &Rpc{Context: ctx, Transport: rpc.Transport}
		require.NoError(t, client.AddTransportFromQr(accountID, acfactory.ConfigQr))
		require.NoError(t, client.StopIo(accountID))

		// Stopped I/O cannot finish work; cancellation must bound the wait.
		stoppedCtx, cancelStopped := context.WithTimeout(ctx, 250*time.Millisecond)
		defer cancelStopped()
		stoppedClient := &Rpc{Context: stoppedCtx, Transport: rpc.Transport}
		require.ErrorIs(t, stoppedClient.WaitForAllWorkDone(accountID), context.DeadlineExceeded)

		require.NoError(t, client.StartIo(accountID))
		require.NoError(t, client.WaitForAllWorkDone(accountID))
		connectivity, err := client.GetConnectivity(accountID)
		require.NoError(t, err)
		const connectivityConnected uint32 = 4000
		require.Equal(t, connectivityConnected, connectivity)
		finished, err := client.IsSendingFinished()
		require.NoError(t, err)
		require.True(t, finished)
	})
}

func TestRpc_PinMessages(t *testing.T) {
	t.Parallel()
	acfactory.WithOnlineAccount(func(rpc *Rpc, accountID uint32) {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		rpc = &Rpc{Context: ctx, Transport: rpc.Transport}
		chatID, err := rpc.CreateChatByContactId(accountID, ContactSelf)
		require.NoError(t, err)
		messageID, err := rpc.MiscSendTextMessage(accountID, chatID, "pin test")
		require.NoError(t, err)
		for _, pinned := range []bool{true, false} {
			require.NoError(t, rpc.SetPinnedMessageState(accountID, messageID, pinned))
			message, err := rpc.GetMessage(accountID, messageID)
			require.NoError(t, err)
			require.Equal(t, pinned, message.IsPinned)
			pins, err := rpc.GetPinnedMessages(accountID, chatID)
			require.NoError(t, err)
			if pinned {
				require.Equal(t, []uint32{messageID}, pins)
			} else {
				require.Empty(t, pins)
			}
		}
	})
}
