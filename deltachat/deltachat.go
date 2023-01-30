package deltachat

// Delta Chat accounts manager. This is the root of the API.
type DeltaChat struct {
	rpc *Rpc
}

// Create a new account.
func (dc DeltaChat) AddAccount() (Account, error) {
	var id uint64
	err := dc.rpc.CallResult(&id, "add_account")
	return newAccount(dc.rpc, id), err
}

// Return a list of all available accounts.
func (dc DeltaChat) Accounts() ([]Account, error) {
	var ids []uint64
	err := dc.rpc.CallResult(&ids, "get_all_account_ids")
	var accounts []Account
	if err == nil {
		accounts = make([]Account, len(ids))
		for i := range ids {
			accounts[i] = newAccount(dc.rpc, ids[i])
		}
	}
	return accounts, err
}

// Start the I/O of all accounts.
func (dc DeltaChat) StartIO() error {
	return dc.rpc.Call("start_io_for_all_accounts") 
}

// Stop the I/O of all accounts.
func (dc DeltaChat) StopIO() error {
	return dc.rpc.Call("stop_io_for_all_accounts") 
}

// Indicate that the network likely has come back or just that the network conditions might have changed.
func (dc DeltaChat) MaybeNetwork() error {
	return dc.rpc.Call("maybe_network") 
}

// Get information about the Delta Chat core in this system.
func (dc DeltaChat) GetSystemInfo() (map[string]any, error) {
	var info map[string]any
	return info, dc.rpc.CallResult(&info, "get_system_info")
}

// Set stock translation strings.
func (dc DeltaChat) SetTranslations(translations map[string]string) error {
	return dc.rpc.Call("set_stock_strings", translations) 
}

// DeltaChat factory
func NewDeltaChat(rpc *Rpc) DeltaChat {
	return DeltaChat{rpc}
}
