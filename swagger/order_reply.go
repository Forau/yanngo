package swagger

import ()

type OrderReply struct {
	OrderId     int64  `json:"order_id,omitempty"`
	ResultCode  string `json:"result_code,omitempty"`
	OrderState  string `json:"order_state,omitempty"`
	ActionState string `json:"action_state,omitempty"`
	Message     string `json:"message,omitempty"`
}
