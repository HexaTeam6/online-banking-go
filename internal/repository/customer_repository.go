package repository

import (
	"context"
	"fmt"

	"github.com/abdurrachmanwahed/online-banking/internal/model"
	"github.com/jmoiron/sqlx"
)

// CustomerRepository defines the interface for customer profile data access.
type CustomerRepository interface {
	GetByAccountNo(ctx context.Context, accountNo int64) (*model.Customer, error)
	Create(ctx context.Context, tx *sqlx.Tx, customer *model.Customer) error
	Update(ctx context.Context, customer *model.Customer) error
	ListAll(ctx context.Context, pagination model.Pagination) ([]model.Customer, int64, error)
}

// customerRepository is the sqlx implementation of CustomerRepository.
type customerRepository struct {
	db *sqlx.DB
}

// NewCustomerRepository creates a new CustomerRepository instance.
func NewCustomerRepository(db *sqlx.DB) CustomerRepository {
	return &customerRepository{db: db}
}

// GetByAccountNo retrieves a customer by their account number.
func (r *customerRepository) GetByAccountNo(ctx context.Context, accountNo int64) (*model.Customer, error) {
	var customer model.Customer
	query := `SELECT account_no, full_name, gender, birth_date, mobile, email FROM tbl_customer WHERE account_no = ?`
	err := r.db.GetContext(ctx, &customer, query, accountNo)
	if err != nil {
		return nil, fmt.Errorf("get customer by account_no %d: %w", accountNo, err)
	}
	return &customer, nil
}

// Create inserts a new customer record within the provided transaction.
func (r *customerRepository) Create(ctx context.Context, tx *sqlx.Tx, customer *model.Customer) error {
	query := `INSERT INTO tbl_customer (account_no, full_name, gender, birth_date, mobile, email) VALUES (?, ?, ?, ?, ?, ?)`
	_, err := tx.ExecContext(ctx, query, customer.AccountNo, customer.FullName, customer.Gender, customer.BirthDate, customer.Mobile, customer.Email)
	if err != nil {
		return fmt.Errorf("create customer for account_no %d: %w", customer.AccountNo, err)
	}
	return nil
}

// Update modifies an existing customer record.
func (r *customerRepository) Update(ctx context.Context, customer *model.Customer) error {
	query := `UPDATE tbl_customer SET full_name = ?, gender = ?, birth_date = ?, mobile = ?, email = ? WHERE account_no = ?`
	result, err := r.db.ExecContext(ctx, query, customer.FullName, customer.Gender, customer.BirthDate, customer.Mobile, customer.Email, customer.AccountNo)
	if err != nil {
		return fmt.Errorf("update customer for account_no %d: %w", customer.AccountNo, err)
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("check rows affected for account_no %d: %w", customer.AccountNo, err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("customer with account_no %d not found", customer.AccountNo)
	}
	return nil
}

// ListAll returns a paginated list of all customers and the total count.
func (r *customerRepository) ListAll(ctx context.Context, pagination model.Pagination) ([]model.Customer, int64, error) {
	var totalCount int64
	countQuery := `SELECT COUNT(*) FROM tbl_customer`
	err := r.db.GetContext(ctx, &totalCount, countQuery)
	if err != nil {
		return nil, 0, fmt.Errorf("count customers: %w", err)
	}

	offset := (pagination.Page - 1) * pagination.PageSize
	query := `SELECT account_no, full_name, gender, birth_date, mobile, email FROM tbl_customer ORDER BY account_no ASC LIMIT ? OFFSET ?`
	var customers []model.Customer
	err = r.db.SelectContext(ctx, &customers, query, pagination.PageSize, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("list customers: %w", err)
	}

	return customers, totalCount, nil
}
