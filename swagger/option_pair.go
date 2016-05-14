package swagger

type OptionPair struct {
	StrikePrice    float64    `json:"strike_price,omitempty"`
	ExpirationDate Date       `json:"expiration_date,omitempty"`
	Call           Instrument `json:"call,omitempty"`
	Put            Instrument `json:"put,omitempty"`
}
