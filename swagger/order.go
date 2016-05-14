package swagger

type Order struct {
	Accno               int64               `json:"accno,omitempty"`
	OrderId             int64               `json:"order_id,omitempty"`
	Price               Amount              `json:"price,omitempty"`
	Volume              float64             `json:"volume,omitempty"`
	Tradable            TradableId          `json:"tradable,omitempty"`
	OpenVolume          float64             `json:"open_volume,omitempty"`
	TradedVolume        float64             `json:"traded_volume,omitempty"`
	Side                string              `json:"side,omitempty"`
	Modified            int64               `json:"modified,omitempty"`
	Reference           string              `json:"reference,omitempty"`
	ActivationCondition ActivationCondition `json:"activation_condition,omitempty"`
	PriceCondition      string              `json:"price_condition,omitempty"`
	VolumeCondition     string              `json:"volume_condition,omitempty"`
	Validity            Validity            `json:"validity,omitempty"`
	ActionState         string              `json:"action_state,omitempty"`
	OrderType           string              `json:"order_type,omitempty"`
	OrderState          string              `json:"order_state,omitempty"`
}
