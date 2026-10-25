# Changelog

## Unreleased

### Added

- new Rpc API: `GetChatSecurejoinQrCode`, `importVcardContents`, `makeVcard`
- RPC server 2.62.0 methods: `InitTransports`, `SetPinnedMessageState`,
  `GetPinnedMessages`, `WaitForAllWorkDone`, `GetAppVersion`, and `IsSendingFinished`.
- `AppSource` for client update information, `ContactFreshness` for contact
  activity, `Message.IsPinned`, `MessageLoadResultMessage.IsPinned`, and the
  `MessagePinned` / `MessageUnpinned` system-message types.
- Schema method-set verification, new API and serialization tests, and an exact
  server-version check in the integration test runner.

### Changed

- Breaking: bindings now match `deltachat-rpc-server` 2.62.0. Older server
  versions are not supported by this revision's verification workflow.
- Breaking: `Contact`, `FullChat`, and `ChatListItemFetchResultChatListItem`
  replace `WasSeenRecently` with `Freshness`. Compare against
  `ContactFreshnessRecentlySeen` where the old boolean was used; `Normal` and
  `Old` are distinct states. Freshness is not a verification or trust signal.
- `DeleteTransport` is now the server's supported relay-removal operation.
  The last transport cannot be removed. `ListTransports` lists active relays.
- Binding generation uses a pinned generator revision and stops on failure.
  The minimum Go version is unchanged.

### Removed

- Breaking: RPC methods `GetProviderInfo`, `GetPushState`, `ListTransportsEx`,
  and `SetTransportUnpublished`, which are absent from server 2.62.0.
  Use `ListTransports` and `DeleteTransport` for current relay management;
  the old unpublished-relay behavior is not preserved.
- Breaking: `AccountConfigured.Addr`; `Contact.IsVerified`, `Contact.VerifierId`,
  and `Contact.NameAndAddr`. These fields are no longer supplied by the server;
  do not infer verification from contact activity or construct a profile address
  from the removed field. Use `ListTransports` when relay addresses are needed.
- Breaking: `NotifyState`, `ProviderInfo`, `TransportListEntry`,
  `EventTypeImapMessageMoved`, `QrFprMismatch`, and `QrFprWithoutAddr`, which are
  absent from the 2.62.0 schema.

## v1.2.14

## Added

- `Bot.SetUiConfig()` and `Bot.GetUiConfig()`
- new `MsgType`: `MsgVcard`

### Changed

- breaking: version numbering changed to: `v1.CORE_VERSION`
- breaking: minimum Chatmail core version required: 2.14.0
- breaking: required golang version: 1.25
- breaking: `EventHandler` and `NewMsgHandler` now have an extra parameter "bot"
- breaking: retrieve events via long polling (added to JSON-RPC server in: https://github.com/chatmail/core/pull/4341/)
- dependencies: upgrade jrpc2 to version v1.1.2

### Fixed

- fix `Reactions` type

## v0.17.0

### Added

- `acfactory.MkdirTemp()` to create a new temporary directory

## v0.16.0

### Added

- `acfactory.StopRpc()` to stop easily Account/Bot/AccountManager's Rpc
- `acfactory.RunningBot()` to get a bot that is already running

### Changed

- breaking: `acfactory.OnlineBot()` now returns a bot that is not running yet

## v0.15.0

### Added

- add `acfactory` package


## v0.14.0

### Added

- add `MsgState` type for `MsgSnapshot.State`
- add `Account.ConnectivityHtml()`
- add type for every event type
- add `Bot.IsRunning()`
- add `BotRunningErr` and `RpcRunningErr`

### Changed

- use `deltachat-rpc-server.exe` as executable name for the RPC process
- remove `Account.CreateChat()`
- serialize incoming events


## v0.13.0

### Changed

- Type string enums and flags (#6)
- Use `Timestamp` type for Unix timestamps from Delta Chat core


## v0.12.0

- fix `Message.WebxdcInfo()`
- add `Bot.RemoveEventHandler()`


## v0.11.1

- fix `Account.SearchMessages()`


## v0.11.0

- add `Chat.SearchMessages()` and `Account.SearchMessages()`


## v0.10.0

- set more flexible interface `io.Writer` for `RpcIO.Stderr`


## v0.9.0

- add ProvideBackup(), GetBackupQr(), GetBackupQrSvg() and GetBackup()


## v0.8.0

- add `Chat.SetMuteDuration()`


## v0.7.0

- add MsgSnapshot.ParseMemberAdded() and MsgSnapshot.ParseMemberRemoved()
- add `EVENT_IMAP_INBOX_IDLE` constant


## v0.6.1

- return error in Rpc.Start() if the Rpc is already started
- bug fix: initialize Rpc.closed to true in NewRpcIO()


## v0.6.0

- bug fix in `Account.QrCode()`
- avoid panics when stopping Rpc and closing event channels
- bug fix: move `Connectivity()` from AccountManager to Account
- remove pasphrase argument from Account.ExportSelfKeys() and Account.ImportSelfKeys()
- fix bug: remove FreshMsgCount() from Account
- fix bug in Account.AddDeviceMsg()
- fix bug in Account.QueryChatListItems() and Account.QueryChatListEntries()
- `Chat.QrCode()` now returns `(string, string, error)` instead of `([2]string, err)`


## v0.5.0

- `Account.QrCode()` now returns `(string, string, error)` instead of `([2]string, err)`


## v0.4.0

- fix bug in Account.FreshMsgsInArrivalOrder()


## v0.3.0

- avoid deadlocks calling Bot.Stop() when the bot is already stopped
- fix bug: Bot.Run() never returning even if Bot.Stop() was called


## v0.2.0

- removed Bot.RunForever(), and now Bot.Run() doesn't require a channel as argument, added Bot.Stop()
  to stop processing events


## v0.1.0

- initial release
