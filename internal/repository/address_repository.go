package repository

import (
	"context"
	"fmt"

	"github.com/abdurrachmanwahed/online-banking/internal/model"
	"github.com/jmoiron/sqlx"
)

// AddressRepository defines the interface for customer address data access.
type AddressRepository interface {
	Create(ctx context.Context, tx *sqlx.Tx, address *model.Address) error
	GetByAccountNo(ctx context.Context, accountNo int64) (*model.Address, error)
	Update(ctx context.Context, address *model.Address) error
}

// addressRepository is the sqlx implementation of AddressRepository.
type addressRepository struct {
	db *sqlx.DB
}

// NewAddressRepository creates a new AddressRepository instance.
func NewAddressRepository(db *sqlx.DB) AddressRepository {
	return &addressRepository{db: db}
}

// Create inserts a new address record within the provided transaction.
func (r *addressRepository) Create(ctx context.Context, tx *sqlx.Tx, address *model.Address) error {
	query := `INSERT INTO tbl_address (account_no, home_address, city, state, pincode) VALUES (?, ?, ?, ?, ?)`
	_, err := tx.ExecContext(ctx, query, address.AccountNo, address.HomeAddress, address.City, address.State, address.Pincode)
	if err != nil {
		return fmt.Errorf("create address for account_no %d: %w", address.AccountNo, err)
	}
	return nil
}

// GetByAccountNo retrieves the address for a given account number.
func (r *addressRepository) GetByAccountNo(ctx context.Context, accountNo int64) (*model.Address, error) {
	var address model.Address
	query := `SELECT account_no, home_address, city, state, pincode FROM tbl_address WHERE account_no = ?`
	err := r.db.GetContext(ctx, &address, query, accountNo)
	if err != nil {
		return nil, fmt.Errorf("get address by account_no %d: %w", accountNo, err)
	}
	return &address, nil
}

// Update modifies an existing address record.
func (r *addressRepository) Update(ctx context.Context, address *model.Address) error {
	query := `UPDATE tbl_address SET home_address = ?, city = ?, state = ?, pincode = ? WHERE account_no = ?`
	result, err := r.db.ExecContext(ctx, query, address.HomeAddress, address.City, address.State, address.Pincode, address.AccountNo)
	if err != nil {
		return fmt.Errorf("update address for account_no %d: %w", address.AccountNo, err)
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("check rows affected for account_no %d: %w", address.AccountNo, err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("address with account_no %d not found", address.AccountNo)
	}
	return nil
}
