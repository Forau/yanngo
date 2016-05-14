package swagger

type Ledger struct {
	Currency      string `json:"currency,omitempty"`
	AccountSum    Amount `json:"account_sum,omitempty"`
	AccountSumAcc Amount `json:"account_sum_acc,omitempty"`
	AccIntDeb     Amount `json:"acc_int_deb,omitempty"`
	AccIntCred    Amount `json:"acc_int_cred,omitempty"`
	ExchangeRate  Amount `json:"exchange_rate,omitempty"`
}
