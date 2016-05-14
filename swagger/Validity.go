package swagger

import ()

type Validity struct {
	Type_      string `json:"type,omitempty"`
	ValidUntil int64  `json:"valid_until,omitempty"`
}
