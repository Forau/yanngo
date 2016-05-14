package swagger

type Validity struct {
	Typ        string `json:"type,omitempty"`
	ValidUntil int64  `json:"valid_until,omitempty"`
}
