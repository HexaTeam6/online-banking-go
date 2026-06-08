package model

// Customer represents customer profile data.
type Customer struct {
	AccountNo int64  `db:"account_no" json:"account_no"`
	FullName  string `db:"full_name" json:"full_name"`
	Gender    string `db:"gender" json:"gender"`
	BirthDate string `db:"birth_date" json:"birth_date"`
	Mobile    string `db:"mobile" json:"mobile"`
	Email     string `db:"email" json:"email"`
}

// Address represents customer address.
type Address struct {
	AccountNo   int64  `db:"account_no" json:"account_no"`
	HomeAddress string `db:"home_address" json:"home_address"`
	City        string `db:"city" json:"city"`
	State       string `db:"state" json:"state"`
	Pincode     int    `db:"pincode" json:"pincode"`
}
