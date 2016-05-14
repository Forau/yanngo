package swagger

import ()

type Status struct {
	Timestamp     int64  `json:"timestamp,omitempty"`
	ValidVersion  bool   `json:"valid_version,omitempty"`
	SystemRunning bool   `json:"system_running,omitempty"`
	Message       string `json:"message,omitempty"`
}
