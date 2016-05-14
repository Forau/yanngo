package swagger

import ()

type PublicTrade struct {
	BrokerBuying  string  `json:"broker_buying,omitempty"`
	BrokerSelling string  `json:"broker_selling,omitempty"`
	Volume        int64   `json:"volume,omitempty"`
	Price         float64 `json:"price,omitempty"`
	TradeId       string  `json:"trade_id,omitempty"`
	TradeType     string  `json:"trade_type,omitempty"`
}
