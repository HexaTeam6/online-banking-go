package model

// LoginRequest represents a customer or admin login request.
type LoginRequest struct {
	Username string `json:"username" validate:"required,min=3,max=64"`
	Password string `json:"password" validate:"required,min=8,max=128"`
}

// LoginResponse is returned on successful authentication.
type LoginResponse struct {
	Token     string `json:"token"`
	ExpiresAt int64  `json:"expires_at"`
}

// RegisterRequest represents a new account registration request.
type RegisterRequest struct {
	FirstName   string `json:"first_name" validate:"required,min=1,max=50"`
	LastName    string `json:"last_name" validate:"required,min=1,max=50"`
	Gender      string `json:"gender" validate:"required,oneof=Male Female Other"`
	BirthDate   string `json:"birth_date" validate:"required"`
	Mobile      string `json:"mobile" validate:"required,min=10,max=15,numeric"`
	Email       string `json:"email" validate:"required,email,max=100"`
	Address     string `json:"address" validate:"required,min=1,max=100"`
	City        string `json:"city" validate:"required,min=1,max=25"`
	State       string `json:"state" validate:"required,min=1,max=25"`
	ZipCode     string `json:"zip_code" validate:"required,numeric,min=1,max=6"`
	Username    string `json:"username" validate:"required,min=3,max=30,alphanum"`
	Password    string `json:"password" validate:"required,min=8,max=128"`
	AccountType string `json:"account_type" validate:"required,oneof=SAVING CURRENT"`
}

// TransferRequest represents a quick transfer request between accounts.
type TransferRequest struct {
	ToAccount int64  `json:"to_account" validate:"required"`
	Amount    int64  `json:"amount" validate:"required,min=500,max=20000"`
	Purpose   string `json:"purpose" validate:"required,min=1,max=100"`
}

// CreateMoneyRequest represents a money request creation.
type CreateMoneyRequest struct {
	ToAccount int64  `json:"to_account" validate:"required"`
	Amount    int64  `json:"amount" validate:"required,min=500,max=20000"`
	Message   string `json:"message" validate:"required,min=1,max=200"`
}

// UpdateProfileRequest represents a partial profile update.
// Pointer fields allow distinguishing between "not provided" and "empty value".
type UpdateProfileRequest struct {
	FullName  *string `json:"full_name,omitempty" validate:"omitempty,min=1,max=100"`
	Gender    *string `json:"gender,omitempty" validate:"omitempty,oneof=Male Female Other"`
	BirthDate *string `json:"birth_date,omitempty"`
	Mobile    *string `json:"mobile,omitempty" validate:"omitempty,min=10,max=15,numeric"`
	Email     *string `json:"email,omitempty" validate:"omitempty,email,max=100"`
	Address   *string `json:"address,omitempty" validate:"omitempty,min=1,max=100"`
	City      *string `json:"city,omitempty" validate:"omitempty,min=1,max=25"`
	State     *string `json:"state,omitempty" validate:"omitempty,min=1,max=25"`
	ZipCode   *string `json:"zip_code,omitempty" validate:"omitempty,numeric,min=1,max=6"`
}

// FeedbackRequest represents a customer feedback submission.
type FeedbackRequest struct {
	Feedback string `json:"feedback" validate:"required,min=1,max=1000"`
	Hearts   int    `json:"hearts" validate:"required,min=1,max=5"`
}

// BalanceAdjustmentRequest represents an admin balance adjustment operation.
type BalanceAdjustmentRequest struct {
	AccountNo int64  `json:"account_no" validate:"required"`
	Operation string `json:"operation" validate:"required,oneof=credit debit"`
	Amount    int64  `json:"amount" validate:"required,min=1,max=99999999999"`
	Purpose   string `json:"purpose" validate:"required,min=1,max=100"`
}

// TransactionSummary holds aggregated transaction statistics.
type TransactionSummary struct {
	TransactionCount int64 `db:"transaction_count" json:"transaction_count"`
	TotalCredit      int64 `db:"total_credit" json:"total_credit"`
	TotalDebit       int64 `db:"total_debit" json:"total_debit"`
}

// DashboardResponse contains dashboard data for a customer account.
type DashboardResponse struct {
	AllTime        TransactionSummary `json:"all_time"`
	CurrentMonth   TransactionSummary `json:"current_month"`
	CurrentBalance int64              `json:"current_balance"`
}
