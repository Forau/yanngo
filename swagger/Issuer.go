package swagger

import ()

type Issuer struct {
	Name     string `json:"name,omitempty"`
	IssuerId int64  `json:"issuer_id,omitempty"`
}
