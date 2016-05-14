package swagger

import ()

type ActivationCondition struct {
	Typ              string  `json:"type,omitempty"`
	TrailingValue    float64 `json:"trailing_value,omitempty"`
	TriggerValue     float64 `json:"trigger_value,omitempty"`
	TriggerCondition string  `json:"trigger_condition,omitempty"`
}
