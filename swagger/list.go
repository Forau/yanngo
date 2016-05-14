package swagger

type List struct {
	Symbol       string `json:"symbol,omitempty"`
	DisplayOrder int64  `json:"display_order,omitempty"`
	ListId       int64  `json:"list_id,omitempty"`
	Name         string `json:"name,omitempty"`
	Country      string `json:"country,omitempty"`
	Region       string `json:"region,omitempty"`
}
