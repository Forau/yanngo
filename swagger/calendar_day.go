package swagger

import ()

type CalendarDay struct {
	Date  Date  `json:"date,omitempty"`
	Open  int64 `json:"open,omitempty"`
	Close int64 `json:"close,omitempty"`
}
