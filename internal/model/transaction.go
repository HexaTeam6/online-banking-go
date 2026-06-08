package model

import "time"

// Transaction represents a single transaction record.
type Transaction struct {
	TransID    int64     `db:"trans_id" json:"trans_id"`
	TransDate  time.Time `db:"trans_date" json:"trans_date"`
	Amount     int64     `db:"amount" json:"amount"`
	TransType  string    `db:"trans_type" json:"trans_type"` // CREDIT or DEBIT
	Purpose    string    `db:"purpose" json:"purpose"`
	ToAccount  int64     `db:"to_account" json:"to_account"`
	AccountNo  int64     `db:"account_no" json:"account_no"`
	AccountBal int64     `db:"account_bal" json:"account_bal"`
}
