package swagger

import ()

type Feed struct {
	Hostname  string `json:"hostname,omitempty"`
	Port      int64  `json:"port,omitempty"`
	Encrypted bool   `json:"encrypted,omitempty"`
}
