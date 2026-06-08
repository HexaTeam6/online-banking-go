package model

import "time"

// LoginHistory represents a login session record.
type LoginHistory struct {
	TokenID    int64      `db:"token_id" json:"token_id"`
	AccountNo  int64      `db:"account_no" json:"account_no"`
	LoginTime  time.Time  `db:"login_time" json:"login_time"`
	LogoutTime *time.Time `db:"logout_time" json:"logout_time"` // nil = active session
	IPAddress  string     `db:"ip_address" json:"ip_address"`
}
