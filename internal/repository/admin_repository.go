package repository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/abdurrachmanwahed/online-banking/internal/model"
	"github.com/jmoiron/sqlx"
)

// AdminRepository handles admin data access.
type AdminRepository interface {
	GetByID(ctx context.Context, adminID int64) (*model.Admin, error)
	GetByEmail(ctx context.Context, email string) (*model.Admin, error)
}

// adminRepository is the sqlx implementation of AdminRepository.
type adminRepository struct {
	db *sqlx.DB
}

// NewAdminRepository creates a new AdminRepository backed by sqlx.
func NewAdminRepository(db *sqlx.DB) AdminRepository {
	return &adminRepository{db: db}
}

// GetByID retrieves an admin by their ID.
func (r *adminRepository) GetByID(ctx context.Context, adminID int64) (*model.Admin, error) {
	var admin model.Admin
	query := `SELECT admin_id, full_name, mobile, email, password FROM tbl_admin WHERE admin_id = ?`
	err := r.db.GetContext(ctx, &admin, query, adminID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &admin, nil
}

// GetByEmail retrieves an admin by their email address.
func (r *adminRepository) GetByEmail(ctx context.Context, email string) (*model.Admin, error) {
	var admin model.Admin
	query := `SELECT admin_id, full_name, mobile, email, password FROM tbl_admin WHERE email = ?`
	err := r.db.GetContext(ctx, &admin, query, email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &admin, nil
}
