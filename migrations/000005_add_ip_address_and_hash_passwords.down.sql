-- Migration 000005 (down): Remove ip_address column from login history
--
-- NOTE: Password hashing cannot be reversed. Bcrypt is a one-way function.
-- Once passwords have been hashed, they cannot be reverted to plaintext.
-- Only the schema change (ip_address column) is reversible.

ALTER TABLE `tbl_login_history`
  DROP COLUMN `ip_address`;
