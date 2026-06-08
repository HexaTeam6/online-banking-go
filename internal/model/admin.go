package model

// Admin represents an admin user.
type Admin struct {
	AdminID  int64  `db:"admin_id" json:"admin_id"`
	FullName string `db:"full_name" json:"full_name"`
	Mobile   string `db:"mobile" json:"mobile"`
	Email    string `db:"email" json:"email"`
	Password string `db:"password" json:"-"` // never serialized
}
