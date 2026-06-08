package model

// Account represents a bank account credential record.
type Account struct {
	AccountNo int64  `db:"account_no" json:"account_no"`
	Username  string `db:"username" json:"username"`
	Password  string `db:"password" json:"-"` // never serialized
}

// Balance represents account balance data.
type Balance struct {
	AccountNo   int64  `db:"account_no" json:"account_no"`
	AccountType string `db:"account_type" json:"account_type"`
	Balance     int64  `db:"balance" json:"balance"`
}
