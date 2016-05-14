package swagger

import ()

type UnderlyingInfo struct {
	InstrumentId int64  `json:"instrument_id,omitempty"`
	Symbol       string `json:"symbol,omitempty"`
	IsinCode     string `json:"isin_code,omitempty"`
}
