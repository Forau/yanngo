package swagger

import ()

type Trade struct {
	Accno        int64      `json:"accno,omitempty"`
	OrderId      int64      `json:"order_id,omitempty"`
	TradeId      string     `json:"trade_id,omitempty"`
	Tradable     TradableId `json:"tradable,omitempty"`
	Price        Amount     `json:"price,omitempty"`
	Volume       float64    `json:"volume,omitempty"`
	Side         string     `json:"side,omitempty"`
	Counterparty string     `json:"counterparty,omitempty"`
	Tradetime    int64      `json:"tradetime,omitempty"`
}
