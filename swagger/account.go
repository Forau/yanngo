package swagger

import ()

type Account struct {
	Accno         int64  `json:"accno,omitempty"`
	Typ           string `json:"type,omitempty"`
	IsDefault     bool   `json:"default,omitempty"`
	Alias         string `json:"alias,omitempty"`
	IsBlocked     bool   `json:"is_blocked,omitempty"`
	BlockedReason string `json:"blocked_reason,omitempty"`
}
