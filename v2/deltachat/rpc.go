package deltachat

import (
	"context"
	"encoding/json"
)

// Delta Chat RPC client. This is the root of the API.
type Rpc struct {
	// Context to be used on calls to Transport.CallResult() and Transport.Call()
	Context   context.Context
	Transport RpcTransport
}

// Test function.
func (rpc *Rpc) Sleep(delay float64) error {
	return rpc.Transport.Call(rpc.Context, "sleep", delay)
}

// Checks if an email address is valid.
func (rpc *Rpc) CheckEmailValidity(email string) (bool, error) {
	var result bool
	err := rpc.Transport.CallResult(rpc.Context, &result, "check_email_validity", email)
	return result, err
}

// Returns general system info.
func (rpc *Rpc) GetSystemInfo() (map[string]string, error) {
	var result map[string]string
	err := rpc.Transport.CallResult(rpc.Context, &result, "get_system_info")
	return result, err
}

// Get the next event, and remove it from the event queue.
//
// If no events have happened since the last `get_next_event`
// (i.e. if the event queue is empty), the response will be returned
// only when a new event fires.
//
// Note that if you are using the `BaseDeltaChat` JavaScript class
// or the `Rpc` Python class, this function will be invoked
// by those classes internally and should not be used manually.
func (rpc *Rpc) GetNextEvent() (Event, error) {
	var result Event
	err := rpc.Transport.CallResult(rpc.Context, &result, "get_next_event")
	return result, err
}

// Waits for at least one event and return a batch of events.
func (rpc *Rpc) GetNextEventBatch() ([]Event, error) {
	var result []Event
	err := rpc.Transport.CallResult(rpc.Context, &result, "get_next_event_batch")
	return result, err
}

func (rpc *Rpc) AddAccount() (uint32, error) {
	var result uint32
	err := rpc.Transport.CallResult(rpc.Context, &result, "add_account")
	return result, err
}

// Imports/migrated an existing account from a database path into this account manager.
// Returns the ID of new account.
func (rpc *Rpc) MigrateAccount(pathToDb string) (uint32, error) {
	var result uint32
	err := rpc.Transport.CallResult(rpc.Context, &result, "migrate_account", pathToDb)
	return result, err
}

func (rpc *Rpc) RemoveAccount(accountId uint32) error {
	return rpc.Transport.Call(rpc.Context, "remove_account", accountId)
}

func (rpc *Rpc) GetAllAccountIds() ([]uint32, error) {
	var result []uint32
	err := rpc.Transport.CallResult(rpc.Context, &result, "get_all_account_ids")
	return result, err
}

// Select account in account manager, this saves the last used account to accounts.toml
func (rpc *Rpc) SelectAccount(id uint32) error {
	return rpc.Transport.Call(rpc.Context, "select_account", id)
}

// Get the selected account from the account manager (on startup it is read from accounts.toml)
func (rpc *Rpc) GetSelectedAccountId() (*uint32, error) {
	var result *uint32
	err := rpc.Transport.CallResult(rpc.Context, &result, "get_selected_account_id")
	return result, err
}

// Set the order of accounts.
// The provided list should contain all account IDs in the desired order.
// If an account ID is missing from the list, it will be appended at the end.
// If the list contains non-existent account IDs, they will be ignored.
func (rpc *Rpc) SetAccountsOrder(order []uint32) error {
	return rpc.Transport.Call(rpc.Context, "set_accounts_order", order)
}

// Get a list of all configured accounts.
func (rpc *Rpc) GetAllAccounts() ([]Account, error) {
	var rawList []json.RawMessage
	if err := rpc.Transport.CallResult(rpc.Context, &rawList, "get_all_accounts"); err != nil {
		return nil, err
	}
	result := make([]Account, len(rawList))
	for i, raw := range rawList {
		if err := unmarshalAccount(raw, &result[i]); err != nil {
			return nil, err
		}
	}
	return result, nil
}

// Starts background tasks for all accounts.
func (rpc *Rpc) StartIoForAllAccounts() error {
	return rpc.Transport.Call(rpc.Context, "start_io_for_all_accounts")
}

// Stops background tasks for all accounts.
func (rpc *Rpc) StopIoForAllAccounts() error {
	return rpc.Transport.Call(rpc.Context, "stop_io_for_all_accounts")
}

// Performs a background fetch for all accounts in parallel with a timeout.
//
// The `AccountsBackgroundFetchDone` event is emitted at the end even in case of timeout.
// Process all events until you get this one and you can safely return to the background
// without forgetting to create notifications caused by timing race conditions.
func (rpc *Rpc) BackgroundFetch(timeoutInSeconds float64) error {
	return rpc.Transport.Call(rpc.Context, "background_fetch", timeoutInSeconds)
}

func (rpc *Rpc) StopBackgroundFetch() error {
	return rpc.Transport.Call(rpc.Context, "stop_background_fetch")
}

// Starts background tasks for a single account.
func (rpc *Rpc) StartIo(accountId uint32) error {
	return rpc.Transport.Call(rpc.Context, "start_io", accountId)
}

// Stops background tasks for a single account.
func (rpc *Rpc) StopIo(accountId uint32) error {
	return rpc.Transport.Call(rpc.Context, "stop_io", accountId)
}

// Get top-level info for an account.
func (rpc *Rpc) GetAccountInfo(accountId uint32) (Account, error) {
	var raw json.RawMessage
	if err := rpc.Transport.CallResult(rpc.Context, &raw, "get_account_info", accountId); err != nil {
		return nil, err
	}
	var result Account
	err := unmarshalAccount(raw, &result)
	return result, err
}

// Get the current push notification state.
func (rpc *Rpc) GetPushState(accountId uint32) (NotifyState, error) {
	var result NotifyState
	err := rpc.Transport.CallResult(rpc.Context, &result, "get_push_state", accountId)
	return result, err
}

// Get the combined filesize of an account in bytes
func (rpc *Rpc) GetAccountFileSize(accountId uint32) (uint64, error) {
	var result uint64
	err := rpc.Transport.CallResult(rpc.Context, &result, "get_account_file_size", accountId)
	return result, err
}

// Returns provider for the given domain.
//
// This function looks up domain in offline database.
//
// For compatibility, email address can be passed to this function
// instead of the domain.
func (rpc *Rpc) GetProviderInfo(accountId uint32, email string) (*ProviderInfo, error) {
	var result *ProviderInfo
	err := rpc.Transport.CallResult(rpc.Context, &result, "get_provider_info", accountId, email)
	return result, err
}

// Checks if the context is already configured.
func (rpc *Rpc) IsConfigured(accountId uint32) (bool, error) {
	var result bool
	err := rpc.Transport.CallResult(rpc.Context, &result, "is_configured", accountId)
	return result, err
}

// Get system info for an account.
func (rpc *Rpc) GetInfo(accountId uint32) (map[string]string, error) {
	var result map[string]string
	err := rpc.Transport.CallResult(rpc.Context, &result, "get_info", accountId)
	return result, err
}

// Get storage usage report as formatted string
func (rpc *Rpc) GetStorageUsageReportString(accountId uint32) (string, error) {
	var result string
	err := rpc.Transport.CallResult(rpc.Context, &result, "get_storage_usage_report_string", accountId)
	return result, err
}

// Get the blob dir.
func (rpc *Rpc) GetBlobDir(accountId uint32) (*string, error) {
	var result *string
	err := rpc.Transport.CallResult(rpc.Context, &result, "get_blob_dir", accountId)
	return result, err
}

// If there was an error while the account was opened
// and migrated to the current version,
// then this function returns it.
//
// This function is useful because the key-contacts migration could fail due to bugs
// and then the account will not work properly.
//
// After opening an account, the UI should call this function
// and show the error string if one is returned.
func (rpc *Rpc) GetMigrationError(accountId uint32) (*string, error) {
	var result *string
	err := rpc.Transport.CallResult(rpc.Context, &result, "get_migration_error", accountId)
	return result, err
}

// Copy file to blob dir.
func (rpc *Rpc) CopyToBlobDir(accountId uint32, path string) (string, error) {
	var result string
	err := rpc.Transport.CallResult(rpc.Context, &result, "copy_to_blob_dir", accountId, path)
	return result, err
}

// Sets the given configuration key.
func (rpc *Rpc) SetConfig(accountId uint32, key string, value *string) error {
	return rpc.Transport.Call(rpc.Context, "set_config", accountId, key, value)
}

// Updates a batch of configuration values.
func (rpc *Rpc) BatchSetConfig(accountId uint32, config map[string]*string) error {
	return rpc.Transport.Call(rpc.Context, "batch_set_config", accountId, config)
}

// Set configuration values from a QR code (technically from the URI stored in it).
// Before this function is called, `check_qr()` should be used to get the QR code type.
//
// "DCACCOUNT:" and "DCLOGIN:" QR codes configure the account, but I/O mustn't be started for
// such QR codes, consider using [`Self::add_transport_from_qr`] which also restarts I/O.
func (rpc *Rpc) SetConfigFromQr(accountId uint32, qrContent string) error {
	return rpc.Transport.Call(rpc.Context, "set_config_from_qr", accountId, qrContent)
}

func (rpc *Rpc) CheckQr(accountId uint32, qrContent string) (Qr, error) {
	var raw json.RawMessage
	if err := rpc.Transport.CallResult(rpc.Context, &raw, "check_qr", accountId, qrContent); err != nil {
		return nil, err
	}
	var result Qr
	err := unmarshalQr(raw, &result)
	return result, err
}

// Returns configuration value for the given key.
func (rpc *Rpc) GetConfig(accountId uint32, key string) (*string, error) {
	var result *string
	err := rpc.Transport.CallResult(rpc.Context, &result, "get_config", accountId, key)
	return result, err
}

func (rpc *Rpc) BatchGetConfig(accountId uint32, keys []string) (map[string]*string, error) {
	var result map[string]*string
	err := rpc.Transport.CallResult(rpc.Context, &result, "batch_get_config", accountId, keys)
	return result, err
}

// Returns all `ui.*` config keys that were set by the UI.
func (rpc *Rpc) GetAllUiConfigKeys(accountId uint32) ([]string, error) {
	var result []string
	err := rpc.Transport.CallResult(rpc.Context, &result, "get_all_ui_config_keys", accountId)
	return result, err
}

func (rpc *Rpc) SetStockStrings(strings map[string]string) error {
	return rpc.Transport.Call(rpc.Context, "set_stock_strings", strings)
}

// Configures this account with the currently set parameters.
// Setup the credential config before calling this.
//
// Deprecated as of 2025-02; use `add_transport_from_qr()`
// or `add_or_update_transport()` instead.
func (rpc *Rpc) Configure(accountId uint32) error {
	return rpc.Transport.Call(rpc.Context, "configure", accountId)
}

// Configures a new email account using the provided parameters
// and adds it as a transport.
//
// If the email address is the same as an existing transport,
// then this existing account will be reconfigured instead of a new one being added.
//
// This function stops and starts IO as needed.
//
// Usually it will be enough to only set `addr` and `password`,
// and all the other settings will be autoconfigured.
//
// During configuration, ConfigureProgress events are emitted;
// they indicate a successful configuration as well as errors
// and may be used to create a progress bar.
// This function will return after configuration is finished.
//
// If configuration is successful,
// the working server parameters will be saved
// and used for connecting to the server.
// The parameters entered by the user will be saved separately
// so that they can be prefilled when the user opens the server-configuration screen again.
//
// See also:
// - [Self::is_configured()] to check whether there is
// at least one working transport.
// - [Self::add_transport_from_qr()] to add a transport
// from a server encoded in a QR code.
// - [Self::list_transports()] to get a list of all configured transports.
// - [Self::delete_transport()] to remove a transport.
// - [Self::set_transport_unpublished()] to set whether contacts see this transport.
func (rpc *Rpc) AddOrUpdateTransport(accountId uint32, param EnteredLoginParam) error {
	return rpc.Transport.Call(rpc.Context, "add_or_update_transport", accountId, param)
}

// Deprecated 2025-04. Alias for [Self::add_or_update_transport()].
func (rpc *Rpc) AddTransport(accountId uint32, param EnteredLoginParam) error {
	return rpc.Transport.Call(rpc.Context, "add_transport", accountId, param)
}

// Adds a new email account as a transport
// using the server encoded in the QR code.
// See [Self::add_or_update_transport].
func (rpc *Rpc) AddTransportFromQr(accountId uint32, qr string) error {
	return rpc.Transport.Call(rpc.Context, "add_transport_from_qr", accountId, qr)
}

// Returns the list of all email accounts that are used as a transport in the current profile.
// Use [Self::add_or_update_transport()] to add or change a transport
// and [Self::delete_transport()] to delete a transport.
// Use [Self::list_transports_ex()] to additionally query
// whether the transports are marked as 'unpublished'.
func (rpc *Rpc) ListTransports(accountId uint32) ([]EnteredLoginParam, error) {
	var result []EnteredLoginParam
	err := rpc.Transport.CallResult(rpc.Context, &result, "list_transports", accountId)
	return result, err
}

// Returns the list of all email accounts that are used as a transport in the current profile.
// Use [Self::add_or_update_transport()] to add or change a transport
// and [Self::delete_transport()] to delete a transport.
func (rpc *Rpc) ListTransportsEx(accountId uint32) ([]TransportListEntry, error) {
	var result []TransportListEntry
	err := rpc.Transport.CallResult(rpc.Context, &result, "list_transports_ex", accountId)
	return result, err
}

// Removes the transport with the specified email address
// (i.e. [EnteredLoginParam::addr]).
func (rpc *Rpc) DeleteTransport(accountId uint32, addr string) error {
	return rpc.Transport.Call(rpc.Context, "delete_transport", accountId, addr)
}

// Change whether the transport is unpublished.
//
// Unpublished transports are not advertised to contacts,
// and self-sent messages are not sent there,
// so that we don't cause extra messages to the corresponding inbox,
// but can still receive messages from contacts who don't know our new transport addresses yet.
//
// The default is false, but when the user updates from a version that didn't have this flag,
// existing secondary transports are set to unpublished,
// so that an existing transport address doesn't suddenly get spammed with a lot of messages.
func (rpc *Rpc) SetTransportUnpublished(accountId uint32, addr string, unpublished bool) error {
	return rpc.Transport.Call(rpc.Context, "set_transport_unpublished", accountId, addr, unpublished)
}

// Signal an ongoing process to stop.
func (rpc *Rpc) StopOngoingProcess(accountId uint32) error {
	return rpc.Transport.Call(rpc.Context, "stop_ongoing_process", accountId)
}

func (rpc *Rpc) ExportSelfKeys(accountId uint32, path string, passphrase *string) error {
	return rpc.Transport.Call(rpc.Context, "export_self_keys", accountId, path, passphrase)
}

func (rpc *Rpc) ImportSelfKeys(accountId uint32, path string, passphrase *string) error {
	return rpc.Transport.Call(rpc.Context, "import_self_keys", accountId, path, passphrase)
}

// Returns the message IDs of all _fresh_ messages of any chat.
// Typically used for implementing notification summaries
// or badge counters e.g. on the app icon.
// The list is already sorted and starts with the most recent fresh message.
//
// Messages belonging to muted chats or to the contact requests are not returned;
// these messages should not be notified
// and also badge counters should not include these messages.
//
// To get the number of fresh messages for a single chat, muted or not,
// use `get_fresh_msg_cnt()`.
func (rpc *Rpc) GetFreshMsgs(accountId uint32) ([]uint32, error) {
	var result []uint32
	err := rpc.Transport.CallResult(rpc.Context, &result, "get_fresh_msgs", accountId)
	return result, err
}

// Get the number of _fresh_ messages in a chat.
// Typically used to implement a badge with a number in the chatlist.
//
// If the specified chat is muted,
// the UI should show the badge counter "less obtrusive",
// e.g. using "gray" instead of "red" color.
func (rpc *Rpc) GetFreshMsgCnt(accountId uint32, chatId uint32) (uint, error) {
	var result uint
	err := rpc.Transport.CallResult(rpc.Context, &result, "get_fresh_msg_cnt", accountId, chatId)
	return result, err
}

// (deprecated) Gets messages to be processed by the bot and returns their IDs.
//
// Only messages with database ID higher than `last_msg_id` config value
// are returned. After processing the messages, the bot should
// update `last_msg_id` by calling [`markseen_msgs`]
// or manually updating the value to avoid getting already
// processed messages.
//
// Deprecated 2026-04: This returns the message's id as soon as the first part arrives,
// even if it is not fully downloaded yet.
// The bot needs to wait for the message to be fully downloaded.
// Since this is usually not the desired behavior,
// bots should instead use the #DC_EVENT_INCOMING_MSG / [`types::events::EventType::IncomingMsg`]
// event for getting notified about new messages.
//
// [`markseen_msgs`]: Self::markseen_msgs
func (rpc *Rpc) GetNextMsgs(accountId uint32) ([]uint32, error) {
	var result []uint32
	err := rpc.Transport.CallResult(rpc.Context, &result, "get_next_msgs", accountId)
	return result, err
}

// (deprecated) Waits for messages to be processed by the bot and returns their IDs.
//
// This function is similar to [`get_next_msgs`],
// but waits for internal new message notification before returning.
// New message notification is sent when new message is added to the database,
// on initialization, when I/O is started and when I/O is stopped.
// This allows bots to use `wait_next_msgs` in a loop to process
// old messages after initialization and during the bot runtime.
// To shutdown the bot, stopping I/O can be used to interrupt
// pending or next `wait_next_msgs` call.
//
// Deprecated 2026-04: This returns the message's id as soon as the first part arrives,
// even if it is not fully downloaded yet.
// The bot needs to wait for the message to be fully downloaded.
// Since this is usually not the desired behavior,
// bots should instead use the #DC_EVENT_INCOMING_MSG / [`types::events::EventType::IncomingMsg`]
// event for getting notified about new messages.
//
// [`get_next_msgs`]: Self::get_next_msgs
func (rpc *Rpc) WaitNextMsgs(accountId uint32) ([]uint32, error) {
	var result []uint32
	err := rpc.Transport.CallResult(rpc.Context, &result, "wait_next_msgs", accountId)
	return result, err
}

// Estimate the number of messages that will be deleted
// by the set_config()-options `delete_device_after` or `delete_server_after`.
// This is typically used to show the estimated impact to the user
// before actually enabling deletion of old messages.
func (rpc *Rpc) EstimateAutoDeletionCount(accountId uint32, fromServer bool, seconds int64) (uint, error) {
	var result uint
	err := rpc.Transport.CallResult(rpc.Context, &result, "estimate_auto_deletion_count", accountId, fromServer, seconds)
	return result, err
}

func (rpc *Rpc) GetChatlistEntries(accountId uint32, listFlags *uint32, queryString *string, queryContactId *uint32) ([]uint32, error) {
	var result []uint32
	err := rpc.Transport.CallResult(rpc.Context, &result, "get_chatlist_entries", accountId, listFlags, queryString, queryContactId)
	return result, err
}

// Returns chats similar to the given one.
//
// Experimental API, subject to change without notice.
func (rpc *Rpc) GetSimilarChatIds(accountId uint32, chatId uint32) ([]uint32, error) {
	var result []uint32
	err := rpc.Transport.CallResult(rpc.Context, &result, "get_similar_chat_ids", accountId, chatId)
	return result, err
}

func (rpc *Rpc) GetChatlistItemsByEntries(accountId uint32, entries []uint32) (map[string]ChatListItemFetchResult, error) {
	var rawMap map[string]json.RawMessage
	if err := rpc.Transport.CallResult(rpc.Context, &rawMap, "get_chatlist_items_by_entries", accountId, entries); err != nil {
		return nil, err
	}
	result := make(map[string]ChatListItemFetchResult, len(rawMap))
	for k, raw := range rawMap {
		var val ChatListItemFetchResult
		if err := unmarshalChatListItemFetchResult(raw, &val); err != nil {
			return nil, err
		}
		result[k] = val
	}
	return result, nil
}

func (rpc *Rpc) GetFullChatById(accountId uint32, chatId uint32) (FullChat, error) {
	var result FullChat
	err := rpc.Transport.CallResult(rpc.Context, &result, "get_full_chat_by_id", accountId, chatId)
	return result, err
}

// get basic info about a chat,
// use chatlist_get_full_chat_by_id() instead if you need more information
func (rpc *Rpc) GetBasicChatInfo(accountId uint32, chatId uint32) (BasicChat, error) {
	var result BasicChat
	err := rpc.Transport.CallResult(rpc.Context, &result, "get_basic_chat_info", accountId, chatId)
	return result, err
}

func (rpc *Rpc) AcceptChat(accountId uint32, chatId uint32) error {
	return rpc.Transport.Call(rpc.Context, "accept_chat", accountId, chatId)
}

func (rpc *Rpc) BlockChat(accountId uint32, chatId uint32) error {
	return rpc.Transport.Call(rpc.Context, "block_chat", accountId, chatId)
}

// Delete a chat.
//
// Messages are deleted from the device and the chat database entry is deleted.
// After that, a `MsgsChanged` event is emitted.
// Messages are deleted from the server in background.
//
// Things that are _not done_ implicitly:
//
// - The chat or the contact is **not blocked**, so new messages from the user/the group may appear as a contact request
// and the user may create the chat again.
// - **Groups are not left** - this would
// be unexpected as (1) deleting a normal chat also does not prevent new mails
// from arriving, (2) leaving a group requires sending a message to
// all group members - especially for groups not used for a longer time, this is
// really unexpected when deletion results in contacting all members again,
// (3) only leaving groups is also a valid usecase.
//
// To leave a chat explicitly, use leave_group()
func (rpc *Rpc) DeleteChat(accountId uint32, chatId uint32) error {
	return rpc.Transport.Call(rpc.Context, "delete_chat", accountId, chatId)
}

// Get encryption info for a chat.
// Get a multi-line encryption info, containing encryption preferences of all members.
// Can be used to find out why messages sent to group are not encrypted.
//
// returns Multi-line text
func (rpc *Rpc) GetChatEncryptionInfo(accountId uint32, chatId uint32) (string, error) {
	var result string
	err := rpc.Transport.CallResult(rpc.Context, &result, "get_chat_encryption_info", accountId, chatId)
	return result, err
}

// Get QR code text that will offer a [SecureJoin](https://securejoin.delta.chat/) invitation.
//
// If `chat_id` is a group chat ID, SecureJoin QR code for the group is returned.
// If `chat_id` is unset, setup contact QR code is returned.
func (rpc *Rpc) GetChatSecurejoinQrCode(accountId uint32, chatId *uint32) (string, error) {
	var result string
	err := rpc.Transport.CallResult(rpc.Context, &result, "get_chat_securejoin_qr_code", accountId, chatId)
	return result, err
}

// Get QR code (text and SVG) that will offer a Setup-Contact or Verified-Group invitation.
// The QR code is compatible to the OPENPGP4FPR format
// so that a basic fingerprint comparison also works e.g. with OpenKeychain.
//
// The scanning device will pass the scanned content to `checkQr()` then;
// if `checkQr()` returns `askVerifyContact` or `askVerifyGroup`
// an out-of-band-verification can be joined using `secure_join()`
//
// @deprecated as of 2026-03; use create_qr_svg(get_chat_securejoin_qr_code()) instead.
//
// chat_id: If set to a group-chat-id,
// the Verified-Group-Invite protocol is offered in the QR code;
// works for protected groups as well as for normal groups.
// If not set, the Setup-Contact protocol is offered in the QR code.
// See https://securejoin.delta.chat/ for details about both protocols.
//
// return format: `[code, svg]`
func (rpc *Rpc) GetChatSecurejoinQrCodeSvg(accountId uint32, chatId *uint32) (Pair[string, string], error) {
	var result Pair[string, string]
	err := rpc.Transport.CallResult(rpc.Context, &result, "get_chat_securejoin_qr_code_svg", accountId, chatId)
	return result, err
}

// Continue a Setup-Contact or Verified-Group-Invite protocol
// started on another device with `get_chat_securejoin_qr_code_svg()`.
// This function is typically called when `check_qr()` returns
// type=AskVerifyContact or type=AskVerifyGroup.
//
// The function returns immediately and the handshake runs in background,
// sending and receiving several messages.
// During the handshake, info messages are added to the chat,
// showing progress, success or errors.
//
// Subsequent calls of `secure_join()` will abort previous, unfinished handshakes.
//
// See https://securejoin.delta.chat/ for details about both protocols.
//
// **qr**: The text of the scanned QR code. Typically, the same string as given
// to `check_qr()`.
//
// **returns**: The chat ID of the joined chat, the UI may redirect to the this chat.
// A returned chat ID does not guarantee that the chat is protected or the belonging contact is verified.
func (rpc *Rpc) SecureJoin(accountId uint32, qr string) (uint32, error) {
	var result uint32
	err := rpc.Transport.CallResult(rpc.Context, &result, "secure_join", accountId, qr)
	return result, err
}

// Like `secure_join()`, but allows to pass a source and a UI-path.
// You only need this if your UI has an option to send statistics
// to Delta Chat's developers.
//
// **source**: The source where the QR code came from.
// E.g. a link that was clicked inside or outside Delta Chat,
// the "Paste from Clipboard" action,
// the "Load QR code as image" action,
// or a QR code scan.
//
// **uipath**: Which UI path did the user use to arrive at the QR code screen.
// If the SecurejoinSource was ExternalLink or InternalLink,
// pass `None` here, because the QR code screen wasn't even opened.
// ```
func (rpc *Rpc) SecureJoinWithUxInfo(accountId uint32, qr string, source *SecurejoinSource, uipath *SecurejoinUiPath) (uint32, error) {
	var result uint32
	err := rpc.Transport.CallResult(rpc.Context, &result, "secure_join_with_ux_info", accountId, qr, source, uipath)
	return result, err
}

func (rpc *Rpc) LeaveGroup(accountId uint32, chatId uint32) error {
	return rpc.Transport.Call(rpc.Context, "leave_group", accountId, chatId)
}

// Remove a member from a group.
//
// If the group is already _promoted_ (any message was sent to the group),
// all group members are informed by a special status message that is sent automatically by this function.
//
// Sends out #DC_EVENT_CHAT_MODIFIED and #DC_EVENT_MSGS_CHANGED if a status message was sent.
func (rpc *Rpc) RemoveContactFromChat(accountId uint32, chatId uint32, contactId uint32) error {
	return rpc.Transport.Call(rpc.Context, "remove_contact_from_chat", accountId, chatId, contactId)
}

// Add a member to a group.
//
// If the group is already _promoted_ (any message was sent to the group),
// all group members are informed by a special status message that is sent automatically by this function.
//
// If the group has group protection enabled, only verified contacts can be added to the group.
//
// Sends out #DC_EVENT_CHAT_MODIFIED and #DC_EVENT_MSGS_CHANGED if a status message was sent.
func (rpc *Rpc) AddContactToChat(accountId uint32, chatId uint32, contactId uint32) error {
	return rpc.Transport.Call(rpc.Context, "add_contact_to_chat", accountId, chatId, contactId)
}

// Get the contact IDs belonging to a chat.
//
// - for normal chats, the function always returns exactly one contact,
// DC_CONTACT_ID_SELF is returned only for SELF-chats.
//
// - for group chats all members are returned, DC_CONTACT_ID_SELF is returned
// explicitly as it may happen that oneself gets removed from a still existing
// group
//
// - for broadcast channels, all recipients are returned, DC_CONTACT_ID_SELF is not included
//
// - for mailing lists, the behavior is not documented currently, we will decide on that later.
// for now, the UI should not show the list for mailing lists.
// (we do not know all members and there is not always a global mailing list address,
// so we could return only SELF or the known members; this is not decided yet)
func (rpc *Rpc) GetChatContacts(accountId uint32, chatId uint32) ([]uint32, error) {
	var result []uint32
	err := rpc.Transport.CallResult(rpc.Context, &result, "get_chat_contacts", accountId, chatId)
	return result, err
}

// Returns contact IDs of the past chat members.
func (rpc *Rpc) GetPastChatContacts(accountId uint32, chatId uint32) ([]uint32, error) {
	var result []uint32
	err := rpc.Transport.CallResult(rpc.Context, &result, "get_past_chat_contacts", accountId, chatId)
	return result, err
}

// Create a new encrypted group chat (with key-contacts).
//
// After creation,
// the group has one member with the ID DC_CONTACT_ID_SELF
// and is in _unpromoted_ state.
// This means, you can add or remove members, change the name,
// the group image and so on without messages being sent to all group members.
//
// This changes as soon as the first message is sent to the group members
// and the group becomes _promoted_.
// After that, all changes are synced with all group members
// by sending status message.
//
// To check, if a chat is still unpromoted, you can look at the `is_unpromoted` property of `BasicChat` or `FullChat`.
// This may be useful if you want to show some help for just created groups.
//
// `protect` argument is deprecated as of 2025-10-22 and is left for compatibility.
// Pass `false` here.
func (rpc *Rpc) CreateGroupChat(accountId uint32, name string, protect bool) (uint32, error) {
	var result uint32
	err := rpc.Transport.CallResult(rpc.Context, &result, "create_group_chat", accountId, name, protect)
	return result, err
}

// Create a new unencrypted group chat.
//
// Same as [`Self::create_group_chat`], but the chat is unencrypted and can only have
// address-contacts.
func (rpc *Rpc) CreateGroupChatUnencrypted(accountId uint32, name string) (uint32, error) {
	var result uint32
	err := rpc.Transport.CallResult(rpc.Context, &result, "create_group_chat_unencrypted", accountId, name)
	return result, err
}

// Deprecated 2025-07 in favor of create_broadcast().
func (rpc *Rpc) CreateBroadcastList(accountId uint32) (uint32, error) {
	var result uint32
	err := rpc.Transport.CallResult(rpc.Context, &result, "create_broadcast_list", accountId)
	return result, err
}

// Create a new, outgoing **broadcast channel**
// (called "Channel" in the UI).
//
// Broadcast channels are similar to groups on the sending device,
// however, recipients get the messages in a read-only chat
// and will not see who the other members are.
//
// Called `broadcast` here rather than `channel`,
// because the word "channel" already appears a lot in the code,
// which would make it hard to grep for it.
//
// After creation, the chat contains no recipients and is in _unpromoted_ state;
// see [`CommandApi::create_group_chat`] for more information on the unpromoted state.
//
// Returns the created chat's id.
func (rpc *Rpc) CreateBroadcast(accountId uint32, chatName string) (uint32, error) {
	var result uint32
	err := rpc.Transport.CallResult(rpc.Context, &result, "create_broadcast", accountId, chatName)
	return result, err
}

// Set group name.
//
// If the group is already _promoted_ (any message was sent to the group),
// or if this is a brodacast channel,
// all members are informed by a special status message that is sent automatically by this function.
//
// Sends out #DC_EVENT_CHAT_MODIFIED and #DC_EVENT_MSGS_CHANGED if a status message was sent.
func (rpc *Rpc) SetChatName(accountId uint32, chatId uint32, newName string) error {
	return rpc.Transport.Call(rpc.Context, "set_chat_name", accountId, chatId, newName)
}

// Set group or broadcast channel description.
//
// If the group is already _promoted_ (any message was sent to the group),
// or if this is a brodacast channel,
// all members are informed by a special status message that is sent automatically by this function.
//
// Sends out #DC_EVENT_CHAT_MODIFIED and #DC_EVENT_MSGS_CHANGED if a status message was sent.
//
// See also [`Self::get_chat_description`] / `getChatDescription()`.
func (rpc *Rpc) SetChatDescription(accountId uint32, chatId uint32, description string) error {
	return rpc.Transport.Call(rpc.Context, "set_chat_description", accountId, chatId, description)
}

// Load the chat description from the database.
//
// UIs show this in the profile page of the chat,
// it is settable by [`Self::set_chat_description`] / `setChatDescription()`.
func (rpc *Rpc) GetChatDescription(accountId uint32, chatId uint32) (string, error) {
	var result string
	err := rpc.Transport.CallResult(rpc.Context, &result, "get_chat_description", accountId, chatId)
	return result, err
}

// Set group profile image.
//
// If the group is already _promoted_ (any message was sent to the group),
// or if this is a brodacast channel,
// all members are informed by a special status message that is sent automatically by this function.
//
// Sends out #DC_EVENT_CHAT_MODIFIED and #DC_EVENT_MSGS_CHANGED if a status message was sent.
//
// To find out the profile image of a chat, use dc_chat_get_profile_image()
//
// @param image_path Full path of the image to use as the group image. The image will immediately be copied to the
// `blobdir`; the original image will not be needed anymore.
// If you pass null here, the group image is deleted (for promoted groups, all members are informed about
// this change anyway).
func (rpc *Rpc) SetChatProfileImage(accountId uint32, chatId uint32, imagePath *string) error {
	return rpc.Transport.Call(rpc.Context, "set_chat_profile_image", accountId, chatId, imagePath)
}

func (rpc *Rpc) SetChatVisibility(accountId uint32, chatId uint32, visibility ChatVisibility) error {
	return rpc.Transport.Call(rpc.Context, "set_chat_visibility", accountId, chatId, visibility)
}

func (rpc *Rpc) SetChatEphemeralTimer(accountId uint32, chatId uint32, timer uint32) error {
	return rpc.Transport.Call(rpc.Context, "set_chat_ephemeral_timer", accountId, chatId, timer)
}

func (rpc *Rpc) GetChatEphemeralTimer(accountId uint32, chatId uint32) (uint32, error) {
	var result uint32
	err := rpc.Transport.CallResult(rpc.Context, &result, "get_chat_ephemeral_timer", accountId, chatId)
	return result, err
}

// Add a message to the device-chat.
// Device-messages usually contain update information
// and some hints that are added during the program runs, multi-device etc.
// The device-message may be defined by a label;
// if a message with the same label was added or skipped before,
// the message is not added again, even if the message was deleted in between.
// If needed, the device-chat is created before.
//
// Sends the `MsgsChanged` event on success.
//
// Setting msg to None will prevent the device message with this label from being added in the future.
func (rpc *Rpc) AddDeviceMessage(accountId uint32, label string, msg *MessageData) (*uint32, error) {
	var result *uint32
	err := rpc.Transport.CallResult(rpc.Context, &result, "add_device_message", accountId, label, msg)
	return result, err
}

// Mark all messages in all chats as _noticed_.
// Skips messages from blocked contacts, but does not skip messages in muted chats.
//
// _Noticed_ messages are no longer _fresh_ and do not count as being unseen
// but are still waiting for being marked as "seen" using markseen_msgs()
// (read receipts aren't sent for noticed messages).
//
// Calling this function usually results in the event #DC_EVENT_MSGS_NOTICED.
// See also markseen_msgs().
func (rpc *Rpc) MarknoticedAllChats(accountId uint32) error {
	return rpc.Transport.Call(rpc.Context, "marknoticed_all_chats", accountId)
}

// Mark all messages in a chat as _noticed_.
// _Noticed_ messages are no longer _fresh_ and do not count as being unseen
// but are still waiting for being marked as "seen" using markseen_msgs()
// (read receipts aren't sent for noticed messages).
//
// Calling this function usually results in the event #DC_EVENT_MSGS_NOTICED.
// See also markseen_msgs().
func (rpc *Rpc) MarknoticedChat(accountId uint32, chatId uint32) error {
	return rpc.Transport.Call(rpc.Context, "marknoticed_chat", accountId, chatId)
}

// Marks the last incoming message in the chat as _fresh_.
//
// UI can use this to offer a "mark unread" option,
// so that already noticed chats get a badge counter again.
func (rpc *Rpc) MarkfreshChat(accountId uint32, chatId uint32) error {
	return rpc.Transport.Call(rpc.Context, "markfresh_chat", accountId, chatId)
}

// Returns the message that is immediately followed by the last seen
// message.
// From the point of view of the user this is effectively
// "first unread", but in reality in the database a seen message
// _can_ be followed by a fresh (unseen) message
// if that message has not been individually marked as seen.
func (rpc *Rpc) GetFirstUnreadMessageOfChat(accountId uint32, chatId uint32) (*uint32, error) {
	var result *uint32
	err := rpc.Transport.CallResult(rpc.Context, &result, "get_first_unread_message_of_chat", accountId, chatId)
	return result, err
}

// Set mute duration of a chat.
//
// The UI can then call is_chat_muted() when receiving a new message
// to decide whether it should trigger an notification.
//
// Muted chats should not sound or vibrate
// and should not show a visual notification in the system area.
// Moreover, muted chats should be excluded from global badge counter
// (get_fresh_msgs() skips muted chats therefore)
// and the in-app, per-chat badge counter should use a less obtrusive color.
//
// Sends out #DC_EVENT_CHAT_MODIFIED.
func (rpc *Rpc) SetChatMuteDuration(accountId uint32, chatId uint32, duration MuteDuration) error {
	return rpc.Transport.Call(rpc.Context, "set_chat_mute_duration", accountId, chatId, duration)
}

// Check whether the chat is currently muted (can be changed by set_chat_mute_duration()).
//
// This is available as a standalone function outside of fullchat, because it might be only needed for notification
func (rpc *Rpc) IsChatMuted(accountId uint32, chatId uint32) (bool, error) {
	var result bool
	err := rpc.Transport.CallResult(rpc.Context, &result, "is_chat_muted", accountId, chatId)
	return result, err
}

// Mark messages as presented to the user.
// Typically, UIs call this function on scrolling through the message list,
// when the messages are presented at least for a little moment.
// The concrete action depends on the type of the chat and on the users settings
// (dc_msgs_presented() may be a better name therefore, but well. :)
//
// - For normal chats, the IMAP state is updated, MDN is sent
// (if set_config()-options `mdns_enabled` is set)
// and the internal state is changed to @ref DC_STATE_IN_SEEN to reflect these actions.
//
// - For contact requests, no IMAP or MDNs is done
// and the internal state is not changed therefore.
// See also marknoticed_chat().
//
// Moreover, timer is started for incoming ephemeral messages.
// This also happens for contact requests chats.
//
// This function updates `last_msg_id` configuration value
// to the maximum of the current value and IDs passed to this function.
// Bots which mark messages as seen can rely on this side effect
// to avoid updating `last_msg_id` value manually.
//
// One #DC_EVENT_MSGS_NOTICED event is emitted per modified chat.
func (rpc *Rpc) MarkseenMsgs(accountId uint32, msgIds []uint32) error {
	return rpc.Transport.Call(rpc.Context, "markseen_msgs", accountId, msgIds)
}

// Returns all messages of a particular chat.
//
// * `add_daymarker` - If `true`, add day markers as `DC_MSG_ID_DAYMARKER` to the result,
// e.g. [1234, 1237, 9, 1239]. The day marker timestamp is the midnight one for the
// corresponding (following) day in the local timezone.
func (rpc *Rpc) GetMessageIds(accountId uint32, chatId uint32, infoOnly bool, addDaymarker bool) ([]uint32, error) {
	var result []uint32
	err := rpc.Transport.CallResult(rpc.Context, &result, "get_message_ids", accountId, chatId, infoOnly, addDaymarker)
	return result, err
}

// Checks if the messages with given IDs exist.
//
// Returns IDs of existing messages.
func (rpc *Rpc) GetExistingMsgIds(accountId uint32, msgIds []uint32) ([]uint32, error) {
	var result []uint32
	err := rpc.Transport.CallResult(rpc.Context, &result, "get_existing_msg_ids", accountId, msgIds)
	return result, err
}

func (rpc *Rpc) GetMessageListItems(accountId uint32, chatId uint32, infoOnly bool, addDaymarker bool) ([]MessageListItem, error) {
	var rawList []json.RawMessage
	if err := rpc.Transport.CallResult(rpc.Context, &rawList, "get_message_list_items", accountId, chatId, infoOnly, addDaymarker); err != nil {
		return nil, err
	}
	result := make([]MessageListItem, len(rawList))
	for i, raw := range rawList {
		if err := unmarshalMessageListItem(raw, &result[i]); err != nil {
			return nil, err
		}
	}
	return result, nil
}

func (rpc *Rpc) GetMessage(accountId uint32, msgId uint32) (Message, error) {
	var result Message
	err := rpc.Transport.CallResult(rpc.Context, &result, "get_message", accountId, msgId)
	return result, err
}

func (rpc *Rpc) GetMessageHtml(accountId uint32, messageId uint32) (*string, error) {
	var result *string
	err := rpc.Transport.CallResult(rpc.Context, &result, "get_message_html", accountId, messageId)
	return result, err
}

// get multiple messages in one call,
// if loading one message fails the error is stored in the result object in it's place.
//
// this is the batch variant of [get_message]
func (rpc *Rpc) GetMessages(accountId uint32, messageIds []uint32) (map[string]MessageLoadResult, error) {
	var rawMap map[string]json.RawMessage
	if err := rpc.Transport.CallResult(rpc.Context, &rawMap, "get_messages", accountId, messageIds); err != nil {
		return nil, err
	}
	result := make(map[string]MessageLoadResult, len(rawMap))
	for k, raw := range rawMap {
		var val MessageLoadResult
		if err := unmarshalMessageLoadResult(raw, &val); err != nil {
			return nil, err
		}
		result[k] = val
	}
	return result, nil
}

// Fetch info desktop needs for creating a notification for a message
func (rpc *Rpc) GetMessageNotificationInfo(accountId uint32, messageId uint32) (MessageNotificationInfo, error) {
	var result MessageNotificationInfo
	err := rpc.Transport.CallResult(rpc.Context, &result, "get_message_notification_info", accountId, messageId)
	return result, err
}

// Delete messages. The messages are deleted on the current device and
// on the IMAP server.
func (rpc *Rpc) DeleteMessages(accountId uint32, messageIds []uint32) error {
	return rpc.Transport.Call(rpc.Context, "delete_messages", accountId, messageIds)
}

// Delete messages. The messages are deleted on the current device,
// on the IMAP server and also for all chat members
func (rpc *Rpc) DeleteMessagesForAll(accountId uint32, messageIds []uint32) error {
	return rpc.Transport.Call(rpc.Context, "delete_messages_for_all", accountId, messageIds)
}

// Get an informational text for a single message. The text is multiline and may
// contain e.g. the raw text of the message.
//
// The max. text returned is typically longer (about 100000 characters) than the
// max. text returned by dc_msg_get_text() (about 30000 characters).
func (rpc *Rpc) GetMessageInfo(accountId uint32, messageId uint32) (string, error) {
	var result string
	err := rpc.Transport.CallResult(rpc.Context, &result, "get_message_info", accountId, messageId)
	return result, err
}

// Returns additional information for single message.
func (rpc *Rpc) GetMessageInfoObject(accountId uint32, messageId uint32) (MessageInfo, error) {
	var result MessageInfo
	err := rpc.Transport.CallResult(rpc.Context, &result, "get_message_info_object", accountId, messageId)
	return result, err
}

// Returns count of read receipts on message.
//
// This view count is meant as a feedback measure for the channel owner only.
func (rpc *Rpc) GetMessageReadReceiptCount(accountId uint32, messageId uint32) (uint, error) {
	var result uint
	err := rpc.Transport.CallResult(rpc.Context, &result, "get_message_read_receipt_count", accountId, messageId)
	return result, err
}

// Returns contacts that sent read receipts and the time of reading.
func (rpc *Rpc) GetMessageReadReceipts(accountId uint32, messageId uint32) ([]MessageReadReceipt, error) {
	var result []MessageReadReceipt
	err := rpc.Transport.CallResult(rpc.Context, &result, "get_message_read_receipts", accountId, messageId)
	return result, err
}

// Asks the core to start downloading a message fully.
// This function is typically called when the user hits the "Download" button
// that is shown by the UI in case `download_state` is `'Available'` or `'Failure'`
//
// On success, the @ref DC_MSG "view type of the message" may change
// or the message may be replaced completely by one or more messages with other message IDs.
// That may happen e.g. in cases where the message was encrypted
// and the type could not be determined without fully downloading.
// Downloaded content can be accessed as usual after download.
//
// To reflect these changes a @ref DC_EVENT_MSGS_CHANGED event will be emitted.
func (rpc *Rpc) DownloadFullMessage(accountId uint32, messageId uint32) error {
	return rpc.Transport.Call(rpc.Context, "download_full_message", accountId, messageId)
}

// Search messages containing the given query string.
// Searching can be done globally (chat_id=None) or in a specified chat only (chat_id set).
//
// Global search results are typically displayed using dc_msg_get_summary(), chat
// search results may just highlight the corresponding messages and present a
// prev/next button.
//
// For the global search, the result is limited to 1000 messages,
// this allows an incremental search done fast.
// So, when getting exactly 1000 messages, the result actually may be truncated;
// the UIs may display sth. like "1000+ messages found" in this case.
// The chat search (if chat_id is set) is not limited.
func (rpc *Rpc) SearchMessages(accountId uint32, query string, chatId *uint32) ([]uint32, error) {
	var result []uint32
	err := rpc.Transport.CallResult(rpc.Context, &result, "search_messages", accountId, query, chatId)
	return result, err
}

func (rpc *Rpc) MessageIdsToSearchResults(accountId uint32, messageIds []uint32) (map[string]MessageSearchResult, error) {
	var result map[string]MessageSearchResult
	err := rpc.Transport.CallResult(rpc.Context, &result, "message_ids_to_search_results", accountId, messageIds)
	return result, err
}

func (rpc *Rpc) SaveMsgs(accountId uint32, messageIds []uint32) error {
	return rpc.Transport.Call(rpc.Context, "save_msgs", accountId, messageIds)
}

// Get a single contact options by ID.
func (rpc *Rpc) GetContact(accountId uint32, contactId uint32) (Contact, error) {
	var result Contact
	err := rpc.Transport.CallResult(rpc.Context, &result, "get_contact", accountId, contactId)
	return result, err
}

// Add a single contact as a result of an explicit user action.
//
// This will always create or look up an address-contact,
// i.e. a contact identified by an email address,
// with all messages sent to and from this contact being unencrypted.
// If the user just clicked on an email address,
// you should first check [`Self::lookup_contact_id_by_addr`]/`lookupContactIdByAddr.`,
// and only if there is no contact yet, call this function here.
//
// Returns contact id of the created or existing contact.
func (rpc *Rpc) CreateContact(accountId uint32, email string, name *string) (uint32, error) {
	var result uint32
	err := rpc.Transport.CallResult(rpc.Context, &result, "create_contact", accountId, email, name)
	return result, err
}

// Returns contact id of the created or existing DM chat with that contact
func (rpc *Rpc) CreateChatByContactId(accountId uint32, contactId uint32) (uint32, error) {
	var result uint32
	err := rpc.Transport.CallResult(rpc.Context, &result, "create_chat_by_contact_id", accountId, contactId)
	return result, err
}

func (rpc *Rpc) BlockContact(accountId uint32, contactId uint32) error {
	return rpc.Transport.Call(rpc.Context, "block_contact", accountId, contactId)
}

func (rpc *Rpc) UnblockContact(accountId uint32, contactId uint32) error {
	return rpc.Transport.Call(rpc.Context, "unblock_contact", accountId, contactId)
}

func (rpc *Rpc) GetBlockedContacts(accountId uint32) ([]Contact, error) {
	var result []Contact
	err := rpc.Transport.CallResult(rpc.Context, &result, "get_blocked_contacts", accountId)
	return result, err
}

// Returns ids of known and unblocked contacts.
//
// By default, key-contacts are listed.
//
// * `list_flags` - A combination of flags:
// - `DC_GCL_ADD_SELF` - Add SELF unless filtered by other parameters.
// - `DC_GCL_ADDRESS` - List address-contacts instead of key-contacts.
// * `query` - A string to filter the list.
func (rpc *Rpc) GetContactIds(accountId uint32, listFlags uint32, query *string) ([]uint32, error) {
	var result []uint32
	err := rpc.Transport.CallResult(rpc.Context, &result, "get_contact_ids", accountId, listFlags, query)
	return result, err
}

// Returns known and unblocked contacts.
//
// Formerly called `getContacts2` in Desktop.
// See [`Self::get_contact_ids`] for parameters and more info.
func (rpc *Rpc) GetContacts(accountId uint32, listFlags uint32, query *string) ([]Contact, error) {
	var result []Contact
	err := rpc.Transport.CallResult(rpc.Context, &result, "get_contacts", accountId, listFlags, query)
	return result, err
}

func (rpc *Rpc) GetContactsByIds(accountId uint32, ids []uint32) (map[string]Contact, error) {
	var result map[string]Contact
	err := rpc.Transport.CallResult(rpc.Context, &result, "get_contacts_by_ids", accountId, ids)
	return result, err
}

func (rpc *Rpc) DeleteContact(accountId uint32, contactId uint32) error {
	return rpc.Transport.Call(rpc.Context, "delete_contact", accountId, contactId)
}

// Sets display name for existing contact.
func (rpc *Rpc) ChangeContactName(accountId uint32, contactId uint32, name string) error {
	return rpc.Transport.Call(rpc.Context, "change_contact_name", accountId, contactId, name)
}

// Get encryption info for a contact.
// Get a multi-line encryption info, containing your fingerprint and the
// fingerprint of the contact, used e.g. to compare the fingerprints for a simple out-of-band verification.
func (rpc *Rpc) GetContactEncryptionInfo(accountId uint32, contactId uint32) (string, error) {
	var result string
	err := rpc.Transport.CallResult(rpc.Context, &result, "get_contact_encryption_info", accountId, contactId)
	return result, err
}

// Looks up a known and unblocked contact with a given e-mail address.
// To get a list of all known and unblocked contacts, use contacts_get_contacts().
//
// **POTENTIAL SECURITY ISSUE**: If there are multiple contacts with this address
// (e.g. an address-contact and a key-contact),
// this looks up the most recently seen contact,
// i.e. which contact is returned depends on which contact last sent a message.
// If the user just clicked on a mailto: link, then this is the best thing you can do.
// But **DO NOT** internally represent contacts by their email address
// and do not use this function to look them up;
// otherwise this function will sometimes look up the wrong contact.
// Instead, you should internally represent contacts by their ids.
//
// To validate an e-mail address independently of the contact database
// use check_email_validity().
func (rpc *Rpc) LookupContactIdByAddr(accountId uint32, addr string) (*uint32, error) {
	var result *uint32
	err := rpc.Transport.CallResult(rpc.Context, &result, "lookup_contact_id_by_addr", accountId, addr)
	return result, err
}

// Parses a vCard file located at the given path. Returns contacts in their original order.
func (rpc *Rpc) ParseVcard(path string) ([]VcardContact, error) {
	var result []VcardContact
	err := rpc.Transport.CallResult(rpc.Context, &result, "parse_vcard", path)
	return result, err
}

// Imports contacts from a vCard file located at the given path.
//
// Returns the ids of created/modified contacts in the order they appear in the vCard.
func (rpc *Rpc) ImportVcard(accountId uint32, path string) ([]uint32, error) {
	var result []uint32
	err := rpc.Transport.CallResult(rpc.Context, &result, "import_vcard", accountId, path)
	return result, err
}

// Imports contacts from a vCard.
//
// Returns the ids of created/modified contacts in the order they appear in the vCard.
func (rpc *Rpc) ImportVcardContents(accountId uint32, vcard string) ([]uint32, error) {
	var result []uint32
	err := rpc.Transport.CallResult(rpc.Context, &result, "import_vcard_contents", accountId, vcard)
	return result, err
}

// Returns a vCard containing contacts with the given ids.
func (rpc *Rpc) MakeVcard(accountId uint32, contacts []uint32) (string, error) {
	var result string
	err := rpc.Transport.CallResult(rpc.Context, &result, "make_vcard", accountId, contacts)
	return result, err
}

// Sets vCard containing the given contacts to the message draft.
func (rpc *Rpc) SetDraftVcard(accountId uint32, msgId uint32, contacts []uint32) error {
	return rpc.Transport.Call(rpc.Context, "set_draft_vcard", accountId, msgId, contacts)
}

// Returns the [`ChatId`] for the 1:1 chat with `contact_id` if it exists.
//
// If it does not exist, `None` is returned.
func (rpc *Rpc) GetChatIdByContactId(accountId uint32, contactId uint32) (*uint32, error) {
	var result *uint32
	err := rpc.Transport.CallResult(rpc.Context, &result, "get_chat_id_by_contact_id", accountId, contactId)
	return result, err
}

// Returns all message IDs of the given types in a chat.
// Typically used to show a gallery.
//
// The list is already sorted and starts with the oldest message.
// Clients should not try to re-sort the list as this would be an expensive action
// and would result in inconsistencies between clients.
//
// Setting `chat_id` to `None` (`null` in typescript) means get messages with media
// from any chat of the currently used account.
func (rpc *Rpc) GetChatMedia(accountId uint32, chatId *uint32, messageType Viewtype, orMessageType2 *Viewtype, orMessageType3 *Viewtype) ([]uint32, error) {
	var result []uint32
	err := rpc.Transport.CallResult(rpc.Context, &result, "get_chat_media", accountId, chatId, messageType, orMessageType2, orMessageType3)
	return result, err
}

func (rpc *Rpc) ExportBackup(accountId uint32, destination string, passphrase *string) error {
	return rpc.Transport.Call(rpc.Context, "export_backup", accountId, destination, passphrase)
}

func (rpc *Rpc) ImportBackup(accountId uint32, path string, passphrase *string) error {
	return rpc.Transport.Call(rpc.Context, "import_backup", accountId, path, passphrase)
}

// Offers a backup for remote devices to retrieve.
//
// Can be canceled by stopping the ongoing process.  Success or failure can be tracked
// via the `ImexProgress` event which should either reach `1000` for success or `0` for
// failure.
//
// This **stops IO** while it is running.
//
// Returns once a remote device has retrieved the backup, or is canceled.
func (rpc *Rpc) ProvideBackup(accountId uint32) error {
	return rpc.Transport.Call(rpc.Context, "provide_backup", accountId)
}

// Returns the text of the QR code for the running [`CommandApi::provide_backup`].
//
// This QR code text can be used in [`CommandApi::get_backup`] on a second device to
// retrieve the backup and setup this second device.
//
// This call will block until the QR code is ready,
// even if there is no concurrent call to [`CommandApi::provide_backup`],
// but will fail after 60 seconds to avoid deadlocks.
func (rpc *Rpc) GetBackupQr(accountId uint32) (string, error) {
	var result string
	err := rpc.Transport.CallResult(rpc.Context, &result, "get_backup_qr", accountId)
	return result, err
}

// Returns the rendered QR code for the running [`CommandApi::provide_backup`].
//
// This QR code can be used in [`CommandApi::get_backup`] on a second device to
// retrieve the backup and setup this second device.
//
// This call will block until the QR code is ready,
// even if there is no concurrent call to [`CommandApi::provide_backup`],
// but will fail after 60 seconds to avoid deadlocks.
//
// @deprecated as of 2026-03; use `create_qr_svg(get_backup_qr())` instead.
//
// Returns the QR code rendered as an SVG image.
func (rpc *Rpc) GetBackupQrSvg(accountId uint32) (string, error) {
	var result string
	err := rpc.Transport.CallResult(rpc.Context, &result, "get_backup_qr_svg", accountId)
	return result, err
}

// Renders the given text as a QR code SVG image.
func (rpc *Rpc) CreateQrSvg(text string) (string, error) {
	var result string
	err := rpc.Transport.CallResult(rpc.Context, &result, "create_qr_svg", text)
	return result, err
}

// Gets a backup from a remote provider.
//
// This retrieves the backup from a remote device over the network and imports it into
// the current device.
//
// Can be canceled by stopping the ongoing process.
//
// Do not forget to call start_io on the account after a successful import,
// otherwise it will not connect to the email server.
func (rpc *Rpc) GetBackup(accountId uint32, qrText string) error {
	return rpc.Transport.Call(rpc.Context, "get_backup", accountId, qrText)
}

// Indicate that the network likely has come back.
// or just that the network conditions might have changed
func (rpc *Rpc) MaybeNetwork() error {
	return rpc.Transport.Call(rpc.Context, "maybe_network")
}

// Get the current connectivity, i.e. whether the device is connected to the IMAP server.
// One of:
// - DC_CONNECTIVITY_NOT_CONNECTED (1000-1999): Show e.g. the string "Not connected" or a red dot
// - DC_CONNECTIVITY_CONNECTING (2000-2999): Show e.g. the string "Connecting…" or a yellow dot
// - DC_CONNECTIVITY_WORKING (3000-3999): Show e.g. the string "Getting new messages" or a spinning wheel
// - DC_CONNECTIVITY_CONNECTED (>=4000): Show e.g. the string "Connected" or a green dot
//
// We don't use exact values but ranges here so that we can split up
// states into multiple states in the future.
//
// Meant as a rough overview that can be shown
// e.g. in the title of the main screen.
//
// If the connectivity changes, a #DC_EVENT_CONNECTIVITY_CHANGED will be emitted.
func (rpc *Rpc) GetConnectivity(accountId uint32) (uint32, error) {
	var result uint32
	err := rpc.Transport.CallResult(rpc.Context, &result, "get_connectivity", accountId)
	return result, err
}

// Get an overview of the current connectivity, and possibly more statistics.
// Meant to give the user more insight about the current status than
// the basic connectivity info returned by get_connectivity(); show this
// e.g., if the user taps on said basic connectivity info.
//
// If this page changes, a #DC_EVENT_CONNECTIVITY_CHANGED will be emitted.
//
// This comes as an HTML from the core so that we can easily improve it
// and the improvement instantly reaches all UIs.
func (rpc *Rpc) GetConnectivityHtml(accountId uint32) (string, error) {
	var result string
	err := rpc.Transport.CallResult(rpc.Context, &result, "get_connectivity_html", accountId)
	return result, err
}

func (rpc *Rpc) GetLocations(accountId uint32, chatId *uint32, contactId *uint32, timestampBegin int64, timestampEnd int64) ([]Location, error) {
	var result []Location
	err := rpc.Transport.CallResult(rpc.Context, &result, "get_locations", accountId, chatId, contactId, timestampBegin, timestampEnd)
	return result, err
}

func (rpc *Rpc) SendWebxdcStatusUpdate(accountId uint32, instanceMsgId uint32, updateStr string, descr *string) error {
	return rpc.Transport.Call(rpc.Context, "send_webxdc_status_update", accountId, instanceMsgId, updateStr, descr)
}

func (rpc *Rpc) SendWebxdcRealtimeData(accountId uint32, instanceMsgId uint32, data []int) error {
	return rpc.Transport.Call(rpc.Context, "send_webxdc_realtime_data", accountId, instanceMsgId, data)
}

func (rpc *Rpc) SendWebxdcRealtimeAdvertisement(accountId uint32, instanceMsgId uint32) error {
	return rpc.Transport.Call(rpc.Context, "send_webxdc_realtime_advertisement", accountId, instanceMsgId)
}

// Leaves the gossip of the webxdc with the given message id.
//
// NB: When this is called before closing a webxdc app in UIs, it must be guaranteed that
// `send_webxdc_realtime_*()` functions aren't called for the given `instance_message_id`
// anymore until the app is open again.
func (rpc *Rpc) LeaveWebxdcRealtime(accountId uint32, instanceMessageId uint32) error {
	return rpc.Transport.Call(rpc.Context, "leave_webxdc_realtime", accountId, instanceMessageId)
}

func (rpc *Rpc) GetWebxdcStatusUpdates(accountId uint32, instanceMsgId uint32, lastKnownSerial uint32) (string, error) {
	var result string
	err := rpc.Transport.CallResult(rpc.Context, &result, "get_webxdc_status_updates", accountId, instanceMsgId, lastKnownSerial)
	return result, err
}

// Get info from a webxdc message
func (rpc *Rpc) GetWebxdcInfo(accountId uint32, instanceMsgId uint32) (WebxdcMessageInfo, error) {
	var result WebxdcMessageInfo
	err := rpc.Transport.CallResult(rpc.Context, &result, "get_webxdc_info", accountId, instanceMsgId)
	return result, err
}

// Get href from a WebxdcInfoMessage which might include a hash holding
// information about a specific position or state in a webxdc app (optional)
func (rpc *Rpc) GetWebxdcHref(accountId uint32, infoMsgId uint32) (*string, error) {
	var result *string
	err := rpc.Transport.CallResult(rpc.Context, &result, "get_webxdc_href", accountId, infoMsgId)
	return result, err
}

// Get blob encoded as base64 from a webxdc message
//
// path is the path of the file within webxdc archive
func (rpc *Rpc) GetWebxdcBlob(accountId uint32, instanceMsgId uint32, path string) (string, error) {
	var result string
	err := rpc.Transport.CallResult(rpc.Context, &result, "get_webxdc_blob", accountId, instanceMsgId, path)
	return result, err
}

// Sets Webxdc file as integration.
// `file` is the .xdc to use as Webxdc integration.
func (rpc *Rpc) SetWebxdcIntegration(accountId uint32, filePath string) error {
	return rpc.Transport.Call(rpc.Context, "set_webxdc_integration", accountId, filePath)
}

// Returns Webxdc instance used for optional integrations.
// UI can open the Webxdc as usual.
// Returns `None` if there is no integration; the caller can add one using `set_webxdc_integration` then.
// `integrate_for` is the chat to get the integration for.
func (rpc *Rpc) InitWebxdcIntegration(accountId uint32, chatId *uint32) (*uint32, error) {
	var result *uint32
	err := rpc.Transport.CallResult(rpc.Context, &result, "init_webxdc_integration", accountId, chatId)
	return result, err
}

// Starts an outgoing call.
func (rpc *Rpc) PlaceOutgoingCall(accountId uint32, chatId uint32, placeCallInfo string, hasVideo bool) (uint32, error) {
	var result uint32
	err := rpc.Transport.CallResult(rpc.Context, &result, "place_outgoing_call", accountId, chatId, placeCallInfo, hasVideo)
	return result, err
}

// Accepts an incoming call.
func (rpc *Rpc) AcceptIncomingCall(accountId uint32, msgId uint32, acceptCallInfo string) error {
	return rpc.Transport.Call(rpc.Context, "accept_incoming_call", accountId, msgId, acceptCallInfo)
}

// Ends incoming or outgoing call.
func (rpc *Rpc) EndCall(accountId uint32, msgId uint32) error {
	return rpc.Transport.Call(rpc.Context, "end_call", accountId, msgId)
}

// Returns information about the call.
func (rpc *Rpc) CallInfo(accountId uint32, msgId uint32) (CallInfo, error) {
	var result CallInfo
	err := rpc.Transport.CallResult(rpc.Context, &result, "call_info", accountId, msgId)
	return result, err
}

// Returns JSON with ICE servers, to be used for WebRTC video calls.
func (rpc *Rpc) IceServers(accountId uint32) (string, error) {
	var result string
	err := rpc.Transport.CallResult(rpc.Context, &result, "ice_servers", accountId)
	return result, err
}

// Makes an HTTP GET request and returns a response.
//
// `url` is the HTTP or HTTPS URL.
func (rpc *Rpc) GetHttpResponse(accountId uint32, url string) (HttpResponse, error) {
	var result HttpResponse
	err := rpc.Transport.CallResult(rpc.Context, &result, "get_http_response", accountId, url)
	return result, err
}

// Forward messages to another chat.
//
// All types of messages can be forwarded,
// however, they will be flagged as such (dc_msg_is_forwarded() is set).
//
// Original sender, info-state and webxdc updates are not forwarded on purpose.
func (rpc *Rpc) ForwardMessages(accountId uint32, messageIds []uint32, chatId uint32) error {
	return rpc.Transport.Call(rpc.Context, "forward_messages", accountId, messageIds, chatId)
}

// Forward messages to a chat in another account.
// See [`Self::forward_messages`] for more info.
func (rpc *Rpc) ForwardMessagesToAccount(srcAccountId uint32, srcMessageIds []uint32, dstAccountId uint32, dstChatId uint32) error {
	return rpc.Transport.Call(rpc.Context, "forward_messages_to_account", srcAccountId, srcMessageIds, dstAccountId, dstChatId)
}

// Resend messages and make information available for newly added chat members.
// Resending sends out the original message, however, recipients and webxdc-status may differ.
// Clients that already have the original message can still ignore the resent message as
// they have tracked the state by dedicated updates.
//
// Some messages cannot be resent, eg. info-messages, drafts, already pending messages or messages that are not sent by SELF.
//
// message_ids all message IDs that should be resend. All messages must belong to the same chat.
func (rpc *Rpc) ResendMessages(accountId uint32, messageIds []uint32) error {
	return rpc.Transport.Call(rpc.Context, "resend_messages", accountId, messageIds)
}

func (rpc *Rpc) SendSticker(accountId uint32, chatId uint32, stickerPath string) (uint32, error) {
	var result uint32
	err := rpc.Transport.CallResult(rpc.Context, &result, "send_sticker", accountId, chatId, stickerPath)
	return result, err
}

// Send a reaction to message.
//
// Reaction is a string of emojis separated by spaces. Reaction to a
// single message can be sent multiple times. The last reaction
// received overrides all previously received reactions. It is
// possible to remove all reactions by sending an empty string.
func (rpc *Rpc) SendReaction(accountId uint32, messageId uint32, reaction []string) (uint32, error) {
	var result uint32
	err := rpc.Transport.CallResult(rpc.Context, &result, "send_reaction", accountId, messageId, reaction)
	return result, err
}

// Returns reactions to the message.
func (rpc *Rpc) GetMessageReactions(accountId uint32, messageId uint32) (*Reactions, error) {
	var result *Reactions
	err := rpc.Transport.CallResult(rpc.Context, &result, "get_message_reactions", accountId, messageId)
	return result, err
}

func (rpc *Rpc) SendMsg(accountId uint32, chatId uint32, data MessageData) (uint32, error) {
	var result uint32
	err := rpc.Transport.CallResult(rpc.Context, &result, "send_msg", accountId, chatId, data)
	return result, err
}

func (rpc *Rpc) SendEditRequest(accountId uint32, msgId uint32, newText string) error {
	return rpc.Transport.Call(rpc.Context, "send_edit_request", accountId, msgId, newText)
}

// Checks if messages can be sent to a given chat.
func (rpc *Rpc) CanSend(accountId uint32, chatId uint32) (bool, error) {
	var result bool
	err := rpc.Transport.CallResult(rpc.Context, &result, "can_send", accountId, chatId)
	return result, err
}

// Saves a file copy at the user-provided path.
//
// Fails if file already exists at the provided path.
func (rpc *Rpc) SaveMsgFile(accountId uint32, msgId uint32, path string) error {
	return rpc.Transport.Call(rpc.Context, "save_msg_file", accountId, msgId, path)
}

func (rpc *Rpc) RemoveDraft(accountId uint32, chatId uint32) error {
	return rpc.Transport.Call(rpc.Context, "remove_draft", accountId, chatId)
}

// Get draft for a chat, if any.
func (rpc *Rpc) GetDraft(accountId uint32, chatId uint32) (*Message, error) {
	var result *Message
	err := rpc.Transport.CallResult(rpc.Context, &result, "get_draft", accountId, chatId)
	return result, err
}

func (rpc *Rpc) MiscGetStickerFolder(accountId uint32) (string, error) {
	var result string
	err := rpc.Transport.CallResult(rpc.Context, &result, "misc_get_sticker_folder", accountId)
	return result, err
}

// Saves a sticker to a collection/folder in the account's sticker folder.
func (rpc *Rpc) MiscSaveSticker(accountId uint32, msgId uint32, collection string) error {
	return rpc.Transport.Call(rpc.Context, "misc_save_sticker", accountId, msgId, collection)
}

// for desktop, get stickers from stickers folder,
// grouped by the collection/folder they are in.
func (rpc *Rpc) MiscGetStickers(accountId uint32) (map[string][]string, error) {
	var result map[string][]string
	err := rpc.Transport.CallResult(rpc.Context, &result, "misc_get_stickers", accountId)
	return result, err
}

// Returns the messageid of the sent message
func (rpc *Rpc) MiscSendTextMessage(accountId uint32, chatId uint32, text string) (uint32, error) {
	var result uint32
	err := rpc.Transport.CallResult(rpc.Context, &result, "misc_send_text_message", accountId, chatId, text)
	return result, err
}

// Send a message to a chat.
//
// This function returns after the message has been placed in the sending queue.
// This does not imply that the message was really sent out yet.
// However, from your view, you're done with the message.
// Sooner or later it will find its way.
//
// **Attaching files:**
//
// Pass the file path in the `file` parameter.
// If `file` is not in the blob directory yet,
// it will be copied into the blob directory.
// If you want, you can delete the file immediately after this function returns.
//
// You can also write the attachment directly into the blob directory
// and then pass the path as the `file` parameter;
// this will prevent an unnecessary copying of the file.
//
// In `filename`, you can pass the original name of the file,
// which will then be shown in the UI.
// in this case the current name of `file` on the filesystem will be ignored.
//
// In order to deduplicate files that contain the same data,
// the file will be named `<hash>.<extension>`, e.g. `ce940175885d7b78f7b7e9f1396611f.jpg`.
//
// NOTE:
// - This function will rename the file. To get the new file path, call `get_file()`.
// - The file must not be modified after this function was called.
// - Images etc. will NOT be recoded.
// In order to recode images,
// use `misc_set_draft` and pass `Image` as the viewtype.
func (rpc *Rpc) MiscSendMsg(accountId uint32, chatId uint32, text *string, file *string, filename *string, location *Pair[float64, float64], quotedMessageId *uint32) (Pair[uint32, Message], error) {
	var result Pair[uint32, Message]
	err := rpc.Transport.CallResult(rpc.Context, &result, "misc_send_msg", accountId, chatId, text, file, filename, location, quotedMessageId)
	return result, err
}

func (rpc *Rpc) MiscSetDraft(accountId uint32, chatId uint32, text *string, file *string, filename *string, quotedMessageId *uint32, viewType *Viewtype) error {
	return rpc.Transport.Call(rpc.Context, "misc_set_draft", accountId, chatId, text, file, filename, quotedMessageId, viewType)
}

func (rpc *Rpc) MiscSendDraft(accountId uint32, chatId uint32) (uint32, error) {
	var result uint32
	err := rpc.Transport.CallResult(rpc.Context, &result, "misc_send_draft", accountId, chatId)
	return result, err
}
