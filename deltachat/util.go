package deltachat

import (
	"fmt"
	"regexp"
	"strings"
)

// Get the first existing account or create a new one if there is no existing accounts.
// Useful for single-account bots/clients.
func GetAccount(rpc *Rpc) AccountId {
	accounts, _ := rpc.GetAllAccountIds()
	if len(accounts) == 0 {
		accountId, _ := rpc.AddAccount()
		return accountId
	}
	return accounts[0]
}

// Extract metadata from system message with type SysmsgTypeMemberAddedToGroup.
func ParseMemberAdded(rpc Rpc, accountId AccountId, msg *MsgSnapshot) (actor ContactId, target ContactId, err error) {
	return parseMemberAddRemove(rpc, accountId, msg, "added")
}

// Extract metadata from system message with type SysmsgTypeMemberRemovedFromGroup.
func ParseMemberRemoved(rpc Rpc, accountId AccountId, msg *MsgSnapshot) (actor ContactId, target ContactId, err error) {
	return parseMemberAddRemove(rpc, accountId, msg, "removed")
}

func parseMemberAddRemove(rpc Rpc, accountId AccountId, msg *MsgSnapshot, action string) (actor ContactId, target ContactId, err error) {
	text := strings.ToLower(msg.Text)
	actor = msg.FromId

	regex := regexp.MustCompile(`^member (.+) ` + action + ` by .+\.$`)
	match := regex.FindStringSubmatch(text)
	if len(match) > 0 {
		target, err := extractContact(rpc, accountId, match[1])
		if err != nil {
			return 0, 0, err
		}
		return actor, target, nil
	}

	regex = regexp.MustCompile(`^you ` + action + ` member (.+)\.$`)
	match = regex.FindStringSubmatch(text)
	if len(match) > 0 {
		target, err := extractContact(rpc, accountId, match[1])
		if err != nil {
			return 0, 0, err
		}
		return actor, target, nil
	}

	if action == "removed" {
		regex = regexp.MustCompile(`^group left by .+\.$`)
		match = regex.FindStringSubmatch(text)
		if len(match) > 0 {
			return actor, actor, nil
		}

		regex = regexp.MustCompile(`^you left the group\.$`)
		if regex.MatchString(text) {
			return actor, actor, nil
		}
	}

	return 0, 0, fmt.Errorf("system message does not match")
}

func extractContact(rpc Rpc, accountId AccountId, text string) (ContactId, error) {
	regex := regexp.MustCompile(`^.*\((.+@.+)\)$`)
	match := regex.FindStringSubmatch(text)
	if len(match) > 0 {
		text = match[1]
	}
	opt, err := rpc.LookupContactIdByAddr(accountId, text)
	return opt.UnwrapOr(0), err
}
