-- Migration 000005: Add ip_address column to login history and prepare for password hashing
--
-- This migration adds the ip_address column needed for tracking client IPs on login.
-- It also alters the column to allow NULL for existing records that don't have IP data.
--
-- NOTE ON PASSWORD HASHING:
-- Bcrypt hashing CANNOT be performed in pure SQL. The hashing of existing plaintext
-- passwords must be done via application code (Go migration script).
--
-- The Go application should:
-- 1. Query all rows from tbl_account and tbl_admin where CHAR_LENGTH(password) < 60
-- 2. For each row, hash the plaintext password using bcrypt with cost factor 12
-- 3. Update the row with the new bcrypt hash
-- 4. Skip any password that is already 60 characters (already hashed)
--
-- This approach is implemented in the Go migration tooling at cmd/migrate or
-- can be run as a one-time script before starting the application.

ALTER TABLE `tbl_login_history`
  ADD COLUMN `ip_address` VARCHAR(45) DEFAULT NULL;
