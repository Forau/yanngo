package swagger

type LedgerInformation struct {
	TotalAccIntDeb  Amount   `json:"total_acc_int_deb,omitempty"`
	TotalAccIntCred Amount   `json:"total_acc_int_cred,omitempty"`
	Total           Amount   `json:"total,omitempty"`
	Ledgers         []Ledger `json:"ledgers,omitempty"`
}
