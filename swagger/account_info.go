package swagger

type AccountInfo struct {
	AccountCurrency            string `json:"account_currency,omitempty"`
	AccountCredit              Amount `json:"account_credit,omitempty"`
	AccountSum                 Amount `json:"account_sum,omitempty"`
	Collateral                 Amount `json:"collateral,omitempty"`
	CreditAccountSum           Amount `json:"credit_account_sum,omitempty"`
	ForwardSum                 Amount `json:"forward_sum,omitempty"`
	FutureSum                  Amount `json:"future_sum,omitempty"`
	UnrealizedFutureProfitLoss Amount `json:"unrealized_future_profit_loss,omitempty"`
	FullMarketvalue            Amount `json:"full_marketvalue,omitempty"`
	Interest                   Amount `json:"interest,omitempty"`
	IntradayCredit             Amount `json:"intraday_credit,omitempty"`
	LoanLimit                  Amount `json:"loan_limit,omitempty"`
	OwnCapital                 Amount `json:"own_capital,omitempty"`
	OwnCapitalMorning          Amount `json:"own_capital_morning,omitempty"`
	PawnValue                  Amount `json:"pawn_value,omitempty"`
	TradingPower               Amount `json:"trading_power,omitempty"`
}
