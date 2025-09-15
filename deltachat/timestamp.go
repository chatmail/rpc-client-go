package deltachat

import (
	"encoding/json"
	"fmt"
	"time"
)

type Timestamp struct {
	time.Time
}

// UnmarshalJSON parses a Delta Chat timestamp into a Timestamp type.
func (timestamp *Timestamp) UnmarshalJSON(b []byte) error {
	var timestampInt int64
	err := json.Unmarshal(b, &timestampInt)
	if err != nil {
		return err
	}
	timestamp.Time = time.Unix(timestampInt, 0)
	return nil
}

// MarshalJSON turns Timestamp back into the format expected by Delta Chat core.
func (timestamp Timestamp) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf("%d", timestamp.Unix())), nil
}
