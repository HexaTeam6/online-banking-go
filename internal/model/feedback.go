package model

import "time"

// Feedback represents a customer feedback entry.
type Feedback struct {
	FeedbackID   int64     `db:"feedback_id" json:"feedback_id"`
	AccountNo    int64     `db:"account_no" json:"account_no"`
	FeedbackText string    `db:"feedback" json:"feedback"`
	Hearts       int       `db:"hearts" json:"hearts"`
	Time         time.Time `db:"time" json:"time"`
}
