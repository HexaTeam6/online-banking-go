package model

import "time"

// MoneyRequest represents a money request between accounts.
type MoneyRequest struct {
	RequestID   int64     `db:"request_id" json:"request_id"`
	AccountNo   int64     `db:"account_no" json:"account_no"` // requester
	ToAccount   int64     `db:"to_account" json:"to_account"` // target
	Amount      int64     `db:"amount" json:"amount"`
	Message     string    `db:"message" json:"message"`
	HasViewed   bool      `db:"hasViewed" json:"has_viewed"`
	Status      string    `db:"status" json:"status"` // PENDING
	RequestDate time.Time `db:"request_date" json:"request_date"`
}
