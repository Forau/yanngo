package swagger

type LeverageFilter struct {
	Issuers              []Issuer `json:"issuers,omitempty"`
	MarketView           []string `json:"market_view,omitempty"`
	ExpirationDates      []string `json:"expiration_dates,omitempty"`
	InstrumentTypes      []string `json:"instrument_types,omitempty"`
	InstrumentGroupTypes []string `json:"instrument_group_types,omitempty"`
	Currencies           []string `json:"currencies,omitempty"`
	NoOfInstruments      int64    `json:"no_of_instruments,omitempty"`
}
